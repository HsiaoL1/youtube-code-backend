package model

import (
	"youtube-code-backend/pkg/common/types"
)

type ChannelDailyStats struct {
	types.BaseModel
	ChannelID       uint64 `gorm:"uniqueIndex:idx_channel_date" json:"channel_id"`
	Date            string `gorm:"uniqueIndex:idx_channel_date;index" json:"date"`
	Views           int64  `gorm:"default:0" json:"views"`
	WatchTimeSeconds int64 `gorm:"default:0" json:"watch_time_seconds"`
	NewSubscribers  int64  `gorm:"default:0" json:"new_subscribers"`
	LostSubscribers int64  `gorm:"default:0" json:"lost_subscribers"`
	RevenueCents    int64  `gorm:"default:0" json:"revenue_cents"`
}

func (ChannelDailyStats) TableName() string { return "channel_daily_stats" }
