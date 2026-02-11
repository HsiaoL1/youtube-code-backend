package model

import (
	"youtube-code-backend/pkg/common/types"
)

// MessageType defines the type of chat message.
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeSystem   MessageType = "system"
	MessageTypeDonation MessageType = "donation"
)

// MessageStatus defines the status of a chat message.
type MessageStatus string

const (
	MessageStatusActive  MessageStatus = "active"
	MessageStatusDeleted MessageStatus = "deleted"
)

// ChatMessage represents a single message in a chat room.
type ChatMessage struct {
	types.BaseModel
	RoomID  uint64        `gorm:"index;not null" json:"room_id"`
	UserID  uint64        `gorm:"not null" json:"user_id"`
	Type    MessageType   `gorm:"size:20;default:text;not null" json:"type"`
	Content string        `gorm:"type:text;not null" json:"content"`
	Status  MessageStatus `gorm:"size:20;default:active;not null" json:"status"`
}

func (ChatMessage) TableName() string { return "chat_messages" }
