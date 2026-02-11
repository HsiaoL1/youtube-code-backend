package model

import (
	"youtube-code-backend/pkg/common/types"
)

// SearchLive represents a live room entry in the search index.
type SearchLive struct {
	types.BaseModel
	RoomID      uint64 `gorm:"uniqueIndex;not null" json:"room_id"`
	Title       string `gorm:"not null" json:"title"`
	ChannelName string `gorm:"not null" json:"channel_name"`
	Category    string `gorm:"size:100" json:"category"`
	IsLive      bool   `gorm:"default:false;not null" json:"is_live"`
}

func (SearchLive) TableName() string { return "search_live" }
