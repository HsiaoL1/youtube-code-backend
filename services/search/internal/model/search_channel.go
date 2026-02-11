package model

import (
	"youtube-code-backend/pkg/common/types"
)

// SearchChannel represents a channel entry in the search index.
type SearchChannel struct {
	types.BaseModel
	ChannelID   uint64 `gorm:"uniqueIndex;not null" json:"channel_id"`
	Handle      string `gorm:"size:255" json:"handle"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
}

func (SearchChannel) TableName() string { return "search_channels" }
