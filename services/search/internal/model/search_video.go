package model

import (
	"youtube-code-backend/pkg/common/types"
)

// SearchVideo represents a video entry in the search index.
type SearchVideo struct {
	types.BaseModel
	VideoID     uint64 `gorm:"uniqueIndex;not null" json:"video_id"`
	Title       string `gorm:"not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	ChannelName string `gorm:"not null" json:"channel_name"`
	Category    string `gorm:"size:100" json:"category"`
	Tags        string `gorm:"type:text" json:"tags"`
	Visibility  string `gorm:"size:50;default:public;not null" json:"visibility"`
	Status      string `gorm:"size:50;default:ready;not null" json:"status"`
}

func (SearchVideo) TableName() string { return "search_videos" }
