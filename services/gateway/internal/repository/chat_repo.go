package repository

import (
	"time"

	"gorm.io/gorm"
)

type ChatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

type ChatMessageDBRow struct {
	ID        uint64    `gorm:"column:id"`
	RoomID    uint64    `gorm:"column:room_id"`
	UserID    uint64    `gorm:"column:user_id"`
	Content   string    `gorm:"column:content"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// FindByRoomID returns chat messages for a room.
func (r *ChatRepo) FindByRoomID(roomID uint64) ([]ChatMessageDBRow, error) {
	var rows []ChatMessageDBRow
	err := r.db.Raw(`
		SELECT id, room_id, user_id, content, created_at
		FROM chat_messages
		WHERE room_id = ? AND status = 'active' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 100
	`, roomID).Scan(&rows).Error
	return rows, err
}
