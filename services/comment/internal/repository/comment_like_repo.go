package repository

import (
	"youtube-code-backend/services/comment/internal/model"

	"gorm.io/gorm"
)

type CommentLikeRepository struct {
	db *gorm.DB
}

func NewCommentLikeRepository(db *gorm.DB) *CommentLikeRepository {
	return &CommentLikeRepository{db: db}
}

func (r *CommentLikeRepository) FindByCommentAndUser(commentID, userID uint64) (*model.CommentLike, error) {
	var like model.CommentLike
	if err := r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&like).Error; err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *CommentLikeRepository) Create(like *model.CommentLike) error {
	return r.db.Create(like).Error
}

func (r *CommentLikeRepository) Delete(commentID, userID uint64) error {
	return r.db.Unscoped().Where("comment_id = ? AND user_id = ?", commentID, userID).
		Delete(&model.CommentLike{}).Error
}

// FindUserLikedCommentIDs returns the set of comment IDs that a user has liked from a given list.
func (r *CommentLikeRepository) FindUserLikedCommentIDs(userID uint64, commentIDs []uint64) (map[uint64]bool, error) {
	var likes []model.CommentLike
	if err := r.db.Where("user_id = ? AND comment_id IN ?", userID, commentIDs).Find(&likes).Error; err != nil {
		return nil, err
	}
	result := make(map[uint64]bool, len(likes))
	for _, l := range likes {
		result[l.CommentID] = true
	}
	return result, nil
}
