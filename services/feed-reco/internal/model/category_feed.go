package model

import (
	"youtube-code-backend/pkg/common/types"
)

// CategoryFeed stores precomputed per-category video rankings.
type CategoryFeed struct {
	types.BaseModel
	Category string  `gorm:"size:100;not null;index:idx_category_position" json:"category"`
	VideoID  uint64  `gorm:"not null" json:"video_id"`
	Score    float64 `gorm:"not null;default:0" json:"score"`
	Position int     `gorm:"not null;index:idx_category_position" json:"position"`
}

func (CategoryFeed) TableName() string { return "category_feeds" }
