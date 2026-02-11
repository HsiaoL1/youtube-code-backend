package model

import (
	"youtube-code-backend/pkg/common/types"
)

type LiveSessionStats struct {
	types.BaseModel
	SessionID       uint64 `gorm:"uniqueIndex" json:"session_id"`
	PeakViewers     int64  `gorm:"default:0" json:"peak_viewers"`
	AvgViewers      int64  `gorm:"default:0" json:"avg_viewers"`
	TotalMessages   int64  `gorm:"default:0" json:"total_messages"`
	DurationSeconds int64  `gorm:"default:0" json:"duration_seconds"`
	UniqueViewers   int64  `gorm:"default:0" json:"unique_viewers"`
}

func (LiveSessionStats) TableName() string { return "live_session_stats" }
