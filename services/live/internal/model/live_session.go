package model

import (
	"time"

	"youtube-code-backend/pkg/common/types"
)

type LiveSession struct {
	types.BaseModel
	RoomID      uint64     `gorm:"index;not null" json:"room_id"`
	StartedAt   time.Time  `gorm:"not null" json:"started_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	Duration    int64      `gorm:"default:0" json:"duration"`
	PeakViewers int64      `gorm:"default:0" json:"peak_viewers"`
	AvgViewers  int64      `gorm:"default:0" json:"avg_viewers"`
}

func (LiveSession) TableName() string { return "live_sessions" }
