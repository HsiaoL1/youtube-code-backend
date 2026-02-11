package repository

import (
	"youtube-code-backend/services/video/internal/model"

	"gorm.io/gorm"
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) FindByVideoAndUser(videoID, userID uint64) (*model.VideoLike, error) {
	var like model.VideoLike
	if err := r.db.Where("video_id = ? AND user_id = ?", videoID, userID).First(&like).Error; err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *LikeRepository) Upsert(like *model.VideoLike) error {
	return r.db.Where("video_id = ? AND user_id = ?", like.VideoID, like.UserID).
		Assign(model.VideoLike{IsLike: like.IsLike}).
		FirstOrCreate(like).Error
}

func (r *LikeRepository) Delete(videoID, userID uint64) error {
	return r.db.Unscoped().Where("video_id = ? AND user_id = ?", videoID, userID).
		Delete(&model.VideoLike{}).Error
}
