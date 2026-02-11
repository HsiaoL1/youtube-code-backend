package model

import (
	"time"

	"youtube-code-backend/pkg/common/types"
)

// ChatMute represents a muted user in a chat room.
type ChatMute struct {
	types.BaseModel
	RoomID    uint64     `gorm:"index;not null" json:"room_id"`
	UserID    uint64     `gorm:"not null" json:"user_id"`
	Reason    string     `gorm:"size:255" json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func (ChatMute) TableName() string { return "chat_mutes" }
