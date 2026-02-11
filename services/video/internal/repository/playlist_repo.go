package repository

import (
	"youtube-code-backend/services/video/internal/model"

	"gorm.io/gorm"
)

type PlaylistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(playlist *model.Playlist) error {
	return r.db.Create(playlist).Error
}

func (r *PlaylistRepository) FindByID(id uint64) (*model.Playlist, error) {
	var playlist model.Playlist
	if err := r.db.First(&playlist, id).Error; err != nil {
		return nil, err
	}
	return &playlist, nil
}

func (r *PlaylistRepository) FindByIDWithItems(id uint64) (*model.Playlist, error) {
	var playlist model.Playlist
	if err := r.db.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("position ASC")
	}).First(&playlist, id).Error; err != nil {
		return nil, err
	}
	return &playlist, nil
}

func (r *PlaylistRepository) Update(playlist *model.Playlist) error {
	return r.db.Save(playlist).Error
}

func (r *PlaylistRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Playlist{}, id).Error
}

func (r *PlaylistRepository) AddItem(item *model.PlaylistItem) error {
	return r.db.Create(item).Error
}

func (r *PlaylistRepository) RemoveItem(playlistID, videoID uint64) error {
	return r.db.Unscoped().Where("playlist_id = ? AND video_id = ?", playlistID, videoID).
		Delete(&model.PlaylistItem{}).Error
}

func (r *PlaylistRepository) GetMaxPosition(playlistID uint64) (int, error) {
	var max *int
	err := r.db.Model(&model.PlaylistItem{}).
		Where("playlist_id = ?", playlistID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&max).Error
	if err != nil {
		return 0, err
	}
	if max == nil {
		return 0, nil
	}
	return *max, nil
}

func (r *PlaylistRepository) IncrementVideoCount(playlistID uint64, delta int64) error {
	return r.db.Model(&model.Playlist{}).Where("id = ?", playlistID).
		Update("video_count", gorm.Expr("video_count + ?", delta)).Error
}
