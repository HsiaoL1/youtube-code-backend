package repository

import (
	"youtube-code-backend/services/video/internal/model"

	"gorm.io/gorm"
)

type WatchRepository struct {
	db *gorm.DB
}

func NewWatchRepository(db *gorm.DB) *WatchRepository {
	return &WatchRepository{db: db}
}

func (r *WatchRepository) FindByUserAndVideo(userID, videoID uint64) (*model.WatchHistory, error) {
	var wh model.WatchHistory
	if err := r.db.Where("user_id = ? AND video_id = ?", userID, videoID).First(&wh).Error; err != nil {
		return nil, err
	}
	return &wh, nil
}

func (r *WatchRepository) Upsert(wh *model.WatchHistory) error {
	return r.db.Where("user_id = ? AND video_id = ?", wh.UserID, wh.VideoID).
		Assign(model.WatchHistory{Progress: wh.Progress, Duration: wh.Duration}).
		FirstOrCreate(wh).Error
}

func (r *WatchRepository) FindByUser(userID uint64, offset, limit int) ([]model.WatchHistory, int64, error) {
	var items []model.WatchHistory
	var total int64

	query := r.db.Model(&model.WatchHistory{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("updated_at DESC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
