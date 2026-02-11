package repository

import (
	"youtube-code-backend/services/comment/internal/model"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) FindByID(id uint64) (*model.Comment, error) {
	var comment model.Comment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) Update(comment *model.Comment) error {
	return r.db.Save(comment).Error
}

func (r *CommentRepository) FindByVideoID(videoID uint64, offset, limit int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	query := r.db.Model(&model.Comment{}).
		Where("video_id = ? AND parent_id IS NULL AND status = ?", videoID, model.CommentStatusActive)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("is_pinned DESC, created_at DESC").Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (r *CommentRepository) FindReplies(parentID uint64, offset, limit int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	query := r.db.Model(&model.Comment{}).
		Where("parent_id = ? AND status = ?", parentID, model.CommentStatusActive)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

// IncrementCount atomically increments a counter column on the comments table.
func (r *CommentRepository) IncrementCount(commentID uint64, column string, delta int64) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", commentID).
		Update(column, gorm.Expr(column+" + ?", delta)).Error
}
