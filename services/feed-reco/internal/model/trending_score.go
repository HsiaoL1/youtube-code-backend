package model

import (
	"time"

	"youtube-code-backend/pkg/common/types"
)

// TrendingScore stores precomputed trending scores for videos.
type TrendingScore struct {
	types.BaseModel
	VideoID   uint64    `gorm:"uniqueIndex;not null" json:"video_id"`
	Score24h  float64   `gorm:"column:score_24h;not null;default:0" json:"score_24h"`
	Score7d   float64   `gorm:"column:score_7d;not null;default:0" json:"score_7d"`
	Category  string    `gorm:"size:100;index" json:"category"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TrendingScore) TableName() string { return "trending_scores" }
