package model

import (
	"youtube-code-backend/pkg/common/types"
)

type AssetStatus string

const (
	AssetStatusPending    AssetStatus = "pending"
	AssetStatusProcessing AssetStatus = "processing"
	AssetStatusReady      AssetStatus = "ready"
	AssetStatusFailed     AssetStatus = "failed"
)

type MediaAsset struct {
	types.BaseModel
	VideoID     uint64      `gorm:"index;not null" json:"video_id"`
	OriginalURL string      `gorm:"size:500;not null" json:"original_url"`
	Duration    int64       `gorm:"default:0" json:"duration"`
	Width       int         `gorm:"default:0" json:"width"`
	Height      int         `gorm:"default:0" json:"height"`
	Codec       string      `gorm:"size:50" json:"codec"`
	Bitrate     int64       `gorm:"default:0" json:"bitrate"`
	FileSize    int64       `gorm:"default:0" json:"file_size"`
	Status      AssetStatus `gorm:"size:20;default:pending;not null" json:"status"`
}

func (MediaAsset) TableName() string { return "media_assets" }
