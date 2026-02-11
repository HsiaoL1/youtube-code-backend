package repository

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type VideoRepo struct {
	db *gorm.DB
}

func NewVideoRepo(db *gorm.DB) *VideoRepo {
	return &VideoRepo{db: db}
}

// VideoDBRow mirrors the videos table columns we SELECT.
type VideoDBRow struct {
	ID           uint64    `gorm:"column:id"`
	ChannelID    uint64    `gorm:"column:channel_id"`
	Type         string    `gorm:"column:type"`
	Title        string    `gorm:"column:title"`
	Description  string    `gorm:"column:description"`
	Status       string    `gorm:"column:status"`
	Visibility   string    `gorm:"column:visibility"`
	Duration     int64     `gorm:"column:duration"`
	ViewCount    int64     `gorm:"column:view_count"`
	LikeCount    int64     `gorm:"column:like_count"`
	ThumbnailURL string    `gorm:"column:thumbnail_url"`
	Tags         string    `gorm:"column:tags"`
	Category     string    `gorm:"column:category"`
	HlsURL       string    `gorm:"column:hls_url"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

const videoSelectCols = `id, channel_id, type, title, description, status, visibility,
	duration, view_count, like_count, thumbnail_url,
	COALESCE(tags, '') AS tags, COALESCE(category, '') AS category,
	COALESCE(hls_url, '') AS hls_url, created_at`

// FindByID returns a single video.
func (r *VideoRepo) FindByID(id uint64) (*VideoDBRow, error) {
	var row VideoDBRow
	err := r.db.Raw(fmt.Sprintf(`SELECT %s FROM videos WHERE id = ? AND deleted_at IS NULL`, videoSelectCols), id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// FindByIDs returns videos matching the given IDs, preserving order.
func (r *VideoRepo) FindByIDs(ids []uint64) ([]VideoDBRow, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []VideoDBRow
	err := r.db.Raw(fmt.Sprintf(`SELECT %s FROM videos WHERE id IN ? AND deleted_at IS NULL ORDER BY id`, videoSelectCols), ids).Scan(&rows).Error
	return rows, err
}

// FindAll provides a flexible query for videos.
func (r *VideoRepo) FindAll(category, sortBy, videoType, status, visibility string, channelID uint64, limit int) ([]VideoDBRow, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}

	if category != "" && category != "All" {
		where = append(where, "category = ?")
		args = append(args, category)
	}
	if videoType != "" {
		where = append(where, "type = ?")
		args = append(args, videoType)
	}
	if status != "" {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if visibility != "" {
		where = append(where, "visibility = ?")
		args = append(args, visibility)
	}
	if channelID > 0 {
		where = append(where, "channel_id = ?")
		args = append(args, channelID)
	}

	orderBy := "created_at DESC"
	switch sortBy {
	case "views":
		orderBy = "view_count DESC"
	case "latest":
		orderBy = "created_at DESC"
	case "likes":
		orderBy = "like_count DESC"
	}

	if limit <= 0 {
		limit = 50
	}

	q := fmt.Sprintf(`SELECT %s FROM videos WHERE %s ORDER BY %s LIMIT %d`,
		videoSelectCols, strings.Join(where, " AND "), orderBy, limit)

	var rows []VideoDBRow
	err := r.db.Raw(q, args...).Scan(&rows).Error
	return rows, err
}

// FindByChannelIDs returns videos for multiple channels.
func (r *VideoRepo) FindByChannelIDs(channelIDs []uint64) ([]VideoDBRow, error) {
	if len(channelIDs) == 0 {
		return nil, nil
	}
	var rows []VideoDBRow
	err := r.db.Raw(fmt.Sprintf(`SELECT %s FROM videos
		WHERE channel_id IN ? AND status = 'ready' AND visibility = 'public' AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT 50`, videoSelectCols), channelIDs).Scan(&rows).Error
	return rows, err
}

// SearchByTitle does a simple ILIKE search.
func (r *VideoRepo) SearchByTitle(q string, videoType string) ([]VideoDBRow, error) {
	where := []string{"deleted_at IS NULL", "title ILIKE ?"}
	args := []any{"%" + q + "%"}
	if videoType != "" {
		where = append(where, "type = ?")
		args = append(args, videoType)
	}
	query := fmt.Sprintf(`SELECT %s FROM videos WHERE %s ORDER BY view_count DESC LIMIT 50`,
		videoSelectCols, strings.Join(where, " AND "))
	var rows []VideoDBRow
	err := r.db.Raw(query, args...).Scan(&rows).Error
	return rows, err
}

// IncrementLike increments like_count and upserts video_likes.
func (r *VideoRepo) IncrementLike(videoID, userID uint64) (int64, error) {
	var newCount int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Upsert video_like
		if err := tx.Exec(`
			INSERT INTO video_likes (video_id, user_id, is_like, created_at, updated_at)
			VALUES (?, ?, true, NOW(), NOW())
			ON CONFLICT (video_id, user_id) DO UPDATE SET is_like = true, updated_at = NOW()
		`, videoID, userID).Error; err != nil {
			return err
		}
		// Increment counter
		if err := tx.Exec(`UPDATE videos SET like_count = like_count + 1, updated_at = NOW() WHERE id = ?`, videoID).Error; err != nil {
			return err
		}
		return tx.Raw(`SELECT like_count FROM videos WHERE id = ?`, videoID).Scan(&newCount).Error
	})
	return newCount, err
}

// CreateFavorite inserts a video_favorite row.
func (r *VideoRepo) CreateFavorite(videoID, userID uint64) error {
	return r.db.Exec(`
		INSERT INTO video_favorites (video_id, user_id, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())
		ON CONFLICT (video_id, user_id) DO NOTHING
	`, videoID, userID).Error
}

// Create inserts a new video.
func (r *VideoRepo) Create(channelID uint64, title, description, coverURL, videoType, category, visibility, tags, hlsURL string) (uint64, error) {
	var id uint64
	err := r.db.Raw(`
		INSERT INTO videos (channel_id, type, title, description, status, visibility, duration, view_count,
			like_count, dislike_count, comment_count, favorite_count, thumbnail_url, tags, category, hls_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'processing', ?, 0, 0, 0, 0, 0, 0, ?, ?, ?, ?, NOW(), NOW())
		RETURNING id
	`, channelID, videoType, title, description, visibility, coverURL, tags, category, hlsURL).Scan(&id).Error
	return id, err
}

// Update updates video fields.
func (r *VideoRepo) Update(id uint64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	sets := []string{}
	args := []any{}
	for k, v := range fields {
		sets = append(sets, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}
	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)
	q := fmt.Sprintf(`UPDATE videos SET %s WHERE id = ?`, strings.Join(sets, ", "))
	return r.db.Exec(q, args...).Error
}

// CountByStatus counts videos matching a status.
func (r *VideoRepo) CountByStatus(status string) (int64, error) {
	var count int64
	q := `SELECT COUNT(*) FROM videos WHERE deleted_at IS NULL`
	args := []any{}
	if status != "" {
		q += ` AND status = ?`
		args = append(args, status)
	}
	err := r.db.Raw(q, args...).Scan(&count).Error
	return count, err
}

// FindByStatusIn returns videos matching any of the given statuses.
func (r *VideoRepo) FindByStatusIn(statuses []string) ([]VideoDBRow, error) {
	var rows []VideoDBRow
	err := r.db.Raw(fmt.Sprintf(`SELECT %s FROM videos WHERE status IN ? AND deleted_at IS NULL ORDER BY created_at DESC`, videoSelectCols), statuses).Scan(&rows).Error
	return rows, err
}

// SumViewsByChannelID sums view_count for a channel.
func (r *VideoRepo) SumViewsByChannelID(channelID uint64) (int64, error) {
	var total int64
	err := r.db.Raw(`SELECT COALESCE(SUM(view_count), 0) FROM videos WHERE channel_id = ? AND deleted_at IS NULL`, channelID).Scan(&total).Error
	return total, err
}
