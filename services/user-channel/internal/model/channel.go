package model

import (
	"youtube-code-backend/pkg/common/types"
)

type Channel struct {
	types.BaseModel
	UserID          uint64        `gorm:"index;not null" json:"user_id"`
	Handle          string        `gorm:"uniqueIndex;size:100;not null" json:"handle"`
	Name            string        `gorm:"size:200;not null" json:"name"`
	Description     string        `gorm:"size:2000" json:"description"`
	Banner          string        `gorm:"size:500" json:"banner"`
	SubscriberCount int64         `gorm:"default:0;not null" json:"subscriber_count"`
	VideoCount      int64         `gorm:"default:0;not null" json:"video_count"`
	Links           []ChannelLink `gorm:"foreignKey:ChannelID" json:"links,omitempty"`
}

func (Channel) TableName() string { return "channels" }

type ChannelLink struct {
	types.BaseModel
	ChannelID uint64 `gorm:"index;not null" json:"channel_id"`
	Title     string `gorm:"size:100;not null" json:"title"`
	URL       string `gorm:"size:500;not null" json:"url"`
}

func (ChannelLink) TableName() string { return "channel_links" }
