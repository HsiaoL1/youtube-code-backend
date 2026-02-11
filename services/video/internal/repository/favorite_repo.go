package repository

import (
	"youtube-code-backend/services/video/internal/model"

	"gorm.io/gorm"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) FindByVideoAndUser(videoID, userID uint64) (*model.VideoFavorite, error) {
	var fav model.VideoFavorite
	if err := r.db.Where("video_id = ? AND user_id = ?", videoID, userID).First(&fav).Error; err != nil {
		return nil, err
	}
	return &fav, nil
}

func (r *FavoriteRepository) Create(fav *model.VideoFavorite) error {
	return r.db.Create(fav).Error
}

func (r *FavoriteRepository) Delete(videoID, userID uint64) error {
	return r.db.Unscoped().Where("video_id = ? AND user_id = ?", videoID, userID).
		Delete(&model.VideoFavorite{}).Error
}
