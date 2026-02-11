package repository

import (
	"youtube-code-backend/services/live/internal/model"

	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(room *model.LiveRoom) error {
	return r.db.Create(room).Error
}

func (r *RoomRepository) FindByID(id uint64) (*model.LiveRoom, error) {
	var room model.LiveRoom
	if err := r.db.First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) Update(room *model.LiveRoom) error {
	return r.db.Save(room).Error
}

func (r *RoomRepository) SoftDelete(id uint64) error {
	return r.db.Delete(&model.LiveRoom{}, id).Error
}

func (r *RoomRepository) FindByStreamKey(streamKey string) (*model.LiveRoom, error) {
	var room model.LiveRoom
	if err := r.db.Where("stream_key = ?", streamKey).First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) FindLiveRooms(offset, limit int) ([]model.LiveRoom, int64, error) {
	var rooms []model.LiveRoom
	var total int64

	query := r.db.Model(&model.LiveRoom{}).Where("status = ?", model.RoomStatusLive)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("viewer_count DESC").Find(&rooms).Error; err != nil {
		return nil, 0, err
	}
	return rooms, total, nil
}

func (r *RoomRepository) FindByChannelID(channelID uint64, offset, limit int) ([]model.LiveRoom, int64, error) {
	var rooms []model.LiveRoom
	var total int64

	query := r.db.Model(&model.LiveRoom{}).Where("channel_id = ?", channelID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&rooms).Error; err != nil {
		return nil, 0, err
	}
	return rooms, total, nil
}
