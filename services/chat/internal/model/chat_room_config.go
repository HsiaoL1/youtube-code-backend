package model

import (
	"youtube-code-backend/pkg/common/types"
)

// ChatRoomConfig holds configuration settings for a chat room.
type ChatRoomConfig struct {
	types.BaseModel
	RoomID           uint64 `gorm:"uniqueIndex;not null" json:"room_id"`
	SlowModeSeconds  int    `gorm:"default:0" json:"slow_mode_seconds"`
	SubscriberOnly   bool   `gorm:"default:false" json:"subscriber_only"`
	MaxMessageLength int    `gorm:"default:200" json:"max_message_length"`
}

func (ChatRoomConfig) TableName() string { return "chat_room_configs" }
