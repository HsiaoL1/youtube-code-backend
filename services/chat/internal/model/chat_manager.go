package model

import (
	"youtube-code-backend/pkg/common/types"
)

// ChatManager represents a user with management privileges in a chat room.
type ChatManager struct {
	types.BaseModel
	RoomID uint64 `gorm:"uniqueIndex:idx_room_user;not null" json:"room_id"`
	UserID uint64 `gorm:"uniqueIndex:idx_room_user;not null" json:"user_id"`
	Role   string `gorm:"size:20;default:moderator;not null" json:"role"`
}

func (ChatManager) TableName() string { return "chat_managers" }
