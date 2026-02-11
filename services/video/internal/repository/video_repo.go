package repository

import (
	"youtube-code-backend/services/video/internal/model"

	"gorm.io/gorm"
)

type VideoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Create(video *model.Video) error {
	return r.db.Create(video).Error
}

func (r *VideoRepository) FindByID(id uint64) (*model.Video, error) {
	var video model.Video
	if err := r.db.First(&video, id).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *VideoRepository) Update(video *model.Video) error {
	return r.db.Save(video).Error
}

func (r *VideoRepository) SoftDelete(id uint64) error {
	return r.db.Delete(&model.Video{}, id).Error
}

func (r *VideoRepository) FindByChannelID(channelID uint64, offset, limit int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	query := r.db.Model(&model.Video{}).Where("channel_id = ?", channelID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&videos).Error; err != nil {
		return nil, 0, err
	}
	return videos, total, nil
}

// IncrementCount atomically increments a counter column on the videos table.
func (r *VideoRepository) IncrementCount(videoID uint64, column string, delta int64) error {
	return r.db.Model(&model.Video{}).Where("id = ?", videoID).
		Update(column, gorm.Expr(column+" + ?", delta)).Error
}
