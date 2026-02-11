package model

import (
	"youtube-code-backend/pkg/common/types"
)

// ChatKeywordFilter represents a keyword filter rule for a chat room.
type ChatKeywordFilter struct {
	types.BaseModel
	RoomID  uint64 `gorm:"index;not null" json:"room_id"`
	Keyword string `gorm:"size:255;not null" json:"keyword"`
	Action  string `gorm:"size:20;default:block;not null" json:"action"`
}

func (ChatKeywordFilter) TableName() string { return "chat_keyword_filters" }
