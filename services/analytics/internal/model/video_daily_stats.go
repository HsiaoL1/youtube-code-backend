package model

import (
	"youtube-code-backend/pkg/common/types"
)

type VideoDailyStats struct {
	types.BaseModel
	VideoID          uint64 `gorm:"uniqueIndex:idx_video_date" json:"video_id"`
	Date             string `gorm:"uniqueIndex:idx_video_date;index" json:"date"`
	Views            int64  `gorm:"default:0" json:"views"`
	WatchTimeSeconds int64  `gorm:"default:0" json:"watch_time_seconds"`
	Likes            int64  `gorm:"default:0" json:"likes"`
	Dislikes         int64  `gorm:"default:0" json:"dislikes"`
	Shares           int64  `gorm:"default:0" json:"shares"`
	Comments         int64  `gorm:"default:0" json:"comments"`
	UniqueViewers    int64  `gorm:"default:0" json:"unique_viewers"`
}

func (VideoDailyStats) TableName() string { return "video_daily_stats" }
