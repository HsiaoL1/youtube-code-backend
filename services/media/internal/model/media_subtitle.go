package model

import (
	"youtube-code-backend/pkg/common/types"
)

type MediaSubtitle struct {
	types.BaseModel
	VideoID  uint64 `gorm:"index;not null" json:"video_id"`
	Language string `gorm:"size:10;not null" json:"language"`
	Label    string `gorm:"size:100;not null" json:"label"`
	URL      string `gorm:"size:500;not null" json:"url"`
	IsAuto   bool   `gorm:"default:false" json:"is_auto"`
}

func (MediaSubtitle) TableName() string { return "media_subtitles" }
