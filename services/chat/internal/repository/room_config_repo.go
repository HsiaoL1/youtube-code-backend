package repository

import (
	"youtube-code-backend/services/chat/internal/model"

	"gorm.io/gorm"
)

// RoomConfigRepository handles database operations for chat room configs.
type RoomConfigRepository struct {
	db *gorm.DB
}

// NewRoomConfigRepository creates a new RoomConfigRepository.
func NewRoomConfigRepository(db *gorm.DB) *RoomConfigRepository {
	return &RoomConfigRepository{db: db}
}

// FindByRoomID returns the room config for the given room.
func (r *RoomConfigRepository) FindByRoomID(roomID uint64) (*model.ChatRoomConfig, error) {
	var cfg model.ChatRoomConfig
	if err := r.db.Where("room_id = ?", roomID).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Upsert creates or updates a room config for the given room.
func (r *RoomConfigRepository) Upsert(cfg *model.ChatRoomConfig) error {
	var existing model.ChatRoomConfig
	err := r.db.Where("room_id = ?", cfg.RoomID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(cfg).Error
	}
	if err != nil {
		return err
	}
	cfg.ID = existing.ID
	return r.db.Save(cfg).Error
}
