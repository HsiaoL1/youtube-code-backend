package repository

import (
	"time"

	"gorm.io/gorm"
)

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

type CommentDBRow struct {
	ID        uint64    `gorm:"column:id"`
	VideoID   uint64    `gorm:"column:video_id"`
	UserID    uint64    `gorm:"column:user_id"`
	ParentID  *uint64   `gorm:"column:parent_id"`
	Content   string    `gorm:"column:content"`
	LikeCount int64     `gorm:"column:like_count"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// FindByVideoID lists comments for a video.
func (r *CommentRepo) FindByVideoID(videoID uint64) ([]CommentDBRow, error) {
	var rows []CommentDBRow
	err := r.db.Raw(`
		SELECT id, video_id, user_id, parent_id, content, like_count, status, created_at
		FROM comments
		WHERE video_id = ? AND status = 'active' AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, videoID).Scan(&rows).Error
	return rows, err
}

// FindByVideoIDs lists comments for multiple videos.
func (r *CommentRepo) FindByVideoIDs(videoIDs []uint64) ([]CommentDBRow, error) {
	if len(videoIDs) == 0 {
		return nil, nil
	}
	var rows []CommentDBRow
	err := r.db.Raw(`
		SELECT id, video_id, user_id, parent_id, content, like_count, status, created_at
		FROM comments
		WHERE video_id IN ? AND status = 'active' AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, videoIDs).Scan(&rows).Error
	return rows, err
}

// Create inserts a new comment.
func (r *CommentRepo) Create(videoID, userID uint64, content string, parentID *uint64) (uint64, error) {
	var id uint64
	err := r.db.Raw(`
		INSERT INTO comments (video_id, user_id, parent_id, content, like_count, reply_count, is_pinned, is_hearted, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 0, 0, false, false, 'active', NOW(), NOW())
		RETURNING id
	`, videoID, userID, parentID, content).Scan(&id).Error
	return id, err
}

// SoftDelete marks a comment as deleted.
func (r *CommentRepo) SoftDelete(id uint64) error {
	return r.db.Exec(`UPDATE comments SET status = 'deleted', updated_at = NOW() WHERE id = ?`, id).Error
}

// IncrementLike increments like_count and upserts comment_likes.
func (r *CommentRepo) IncrementLike(commentID, userID uint64) (int64, error) {
	var newCount int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			INSERT INTO comment_likes (comment_id, user_id, created_at, updated_at)
			VALUES (?, ?, NOW(), NOW())
			ON CONFLICT (comment_id, user_id) DO NOTHING
		`, commentID, userID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE comments SET like_count = like_count + 1, updated_at = NOW() WHERE id = ?`, commentID).Error; err != nil {
			return err
		}
		return tx.Raw(`SELECT like_count FROM comments WHERE id = ?`, commentID).Scan(&newCount).Error
	})
	return newCount, err
}

// FindByID returns a single comment.
func (r *CommentRepo) FindByID(id uint64) (*CommentDBRow, error) {
	var row CommentDBRow
	err := r.db.Raw(`
		SELECT id, video_id, user_id, parent_id, content, like_count, status, created_at
		FROM comments WHERE id = ? AND deleted_at IS NULL
	`, id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

// CountAll counts all active comments.
func (r *CommentRepo) CountAll() (int64, error) {
	var count int64
	err := r.db.Raw(`SELECT COUNT(*) FROM comments WHERE status = 'active' AND deleted_at IS NULL`).Scan(&count).Error
	return count, err
}
