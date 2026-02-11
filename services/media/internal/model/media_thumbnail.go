package model

import (
	"youtube-code-backend/pkg/common/types"
)

type MediaThumbnail struct {
	types.BaseModel
	VideoID   uint64 `gorm:"index;not null" json:"video_id"`
	URL       string `gorm:"size:500;not null" json:"url"`
	Width     int    `gorm:"default:0" json:"width"`
	Height    int    `gorm:"default:0" json:"height"`
	IsDefault bool   `gorm:"default:false" json:"is_default"`
}

func (MediaThumbnail) TableName() string { return "media_thumbnails" }
