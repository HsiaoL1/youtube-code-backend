package repository

import (
	"youtube-code-backend/services/chat/internal/model"

	"gorm.io/gorm"
)

// MessageRepository handles database operations for chat messages.
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new MessageRepository.
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create inserts a new chat message.
func (r *MessageRepository) Create(msg *model.ChatMessage) error {
	return r.db.Create(msg).Error
}

// FindByRoomID returns paginated messages for a room, ordered by newest first.
func (r *MessageRepository) FindByRoomID(roomID uint64, offset, limit int) ([]model.ChatMessage, int64, error) {
	var messages []model.ChatMessage
	var total int64

	query := r.db.Model(&model.ChatMessage{}).
		Where("room_id = ? AND status = ?", roomID, model.MessageStatusActive)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// SoftDelete marks a message as deleted.
func (r *MessageRepository) SoftDelete(id uint64) error {
	return r.db.Model(&model.ChatMessage{}).Where("id = ?", id).
		Update("status", model.MessageStatusDeleted).Error
}
