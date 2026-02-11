package model

import "youtube-code-backend/pkg/common/types"

type WatchHistory struct {
	types.BaseModel
	UserID   uint64 `gorm:"index;not null" json:"user_id"`
	VideoID  uint64 `gorm:"not null" json:"video_id"`
	Progress int64  `gorm:"default:0" json:"progress"`
	Duration int64  `gorm:"default:0" json:"duration"`
}

func (WatchHistory) TableName() string { return "watch_history" }
