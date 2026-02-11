package repository

import (
	"time"

	"gorm.io/gorm"
)

type LiveRepo struct {
	db *gorm.DB
}

func NewLiveRepo(db *gorm.DB) *LiveRepo {
	return &LiveRepo{db: db}
}

type LiveRoomDBRow struct {
	ID           uint64    `gorm:"column:id"`
	ChannelID    uint64    `gorm:"column:channel_id"`
	Title        string    `gorm:"column:title"`
	Description  string    `gorm:"column:description"`
	Status       string    `gorm:"column:status"`
	PlaybackURL  string    `gorm:"column:playback_url"`
	ViewerCount  int64     `gorm:"column:viewer_count"`
	Category     string    `gorm:"column:category"`
	ThumbnailURL string    `gorm:"column:thumbnail_url"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

// FindLive returns live rooms, optionally filtered by category.
func (r *LiveRepo) FindLive(category string) ([]LiveRoomDBRow, error) {
	q := `SELECT id, channel_id, title, description, status, COALESCE(playback_url, '') AS playback_url,
		viewer_count, COALESCE(category, '') AS category, COALESCE(thumbnail_url, '') AS thumbnail_url, created_at
		FROM live_rooms WHERE status = 'live' AND deleted_at IS NULL`
	args := []any{}
	if category != "" && category != "All" {
		q += ` AND category = ?`
		args = append(args, category)
	}
	q += ` ORDER BY viewer_count DESC`
	var rows []LiveRoomDBRow
	err := r.db.Raw(q, args...).Scan(&rows).Error
	return rows, err
}

// FindByID returns a single live room.
func (r *LiveRepo) FindByID(id uint64) (*LiveRoomDBRow, error) {
	var row LiveRoomDBRow
	err := r.db.Raw(`
		SELECT id, channel_id, title, description, status, COALESCE(playback_url, '') AS playback_url,
		viewer_count, COALESCE(category, '') AS category, COALESCE(thumbnail_url, '') AS thumbnail_url, created_at
		FROM live_rooms WHERE id = ? AND deleted_at IS NULL
	`, id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// FindByChannelID returns the live room for a channel.
func (r *LiveRepo) FindByChannelID(channelID uint64) (*LiveRoomDBRow, error) {
	var row LiveRoomDBRow
	err := r.db.Raw(`
		SELECT id, channel_id, title, description, status, COALESCE(playback_url, '') AS playback_url,
		viewer_count, COALESCE(category, '') AS category, COALESCE(thumbnail_url, '') AS thumbnail_url, created_at
		FROM live_rooms WHERE channel_id = ? AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`, channelID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// ToggleStatus toggles a live room between live and idle/ended.
func (r *LiveRepo) ToggleStatus(channelID uint64) error {
	// Find current status
	var status string
	if err := r.db.Raw(`SELECT status FROM live_rooms WHERE channel_id = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`, channelID).Scan(&status).Error; err != nil {
		return err
	}
	newStatus := "live"
	if status == "live" {
		newStatus = "ended"
	}
	return r.db.Exec(`UPDATE live_rooms SET status = ?, updated_at = NOW() WHERE channel_id = ? AND deleted_at IS NULL`, newStatus, channelID).Error
}

// SearchByTitle searches live rooms by title.
func (r *LiveRepo) SearchByTitle(q string) ([]LiveRoomDBRow, error) {
	var rows []LiveRoomDBRow
	err := r.db.Raw(`
		SELECT id, channel_id, title, description, status, COALESCE(playback_url, '') AS playback_url,
		viewer_count, COALESCE(category, '') AS category, COALESCE(thumbnail_url, '') AS thumbnail_url, created_at
		FROM live_rooms WHERE title ILIKE ? AND deleted_at IS NULL ORDER BY viewer_count DESC LIMIT 20
	`, "%"+q+"%").Scan(&rows).Error
	return rows, err
}
