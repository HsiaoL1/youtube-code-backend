package repository

import (
	"time"

	"youtube-code-backend/services/chat/internal/model"

	"gorm.io/gorm"
)

// MuteRepository handles database operations for chat mutes.
type MuteRepository struct {
	db *gorm.DB
}

// NewMuteRepository creates a new MuteRepository.
func NewMuteRepository(db *gorm.DB) *MuteRepository {
	return &MuteRepository{db: db}
}

// Create inserts a new chat mute record.
func (r *MuteRepository) Create(mute *model.ChatMute) error {
	return r.db.Create(mute).Error
}

// FindActiveByRoomAndUser finds an active (non-expired) mute for a user in a room.
func (r *MuteRepository) FindActiveByRoomAndUser(roomID, userID uint64) (*model.ChatMute, error) {
	var mute model.ChatMute
	now := time.Now()
	if err := r.db.Where("room_id = ? AND user_id = ? AND (expires_at IS NULL OR expires_at > ?)", roomID, userID, now).
		First(&mute).Error; err != nil {
		return nil, err
	}
	return &mute, nil
}

// DeleteByRoomAndUser removes mutes for a user in a room (hard delete).
func (r *MuteRepository) DeleteByRoomAndUser(roomID, userID uint64) error {
	result := r.db.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&model.ChatMute{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
