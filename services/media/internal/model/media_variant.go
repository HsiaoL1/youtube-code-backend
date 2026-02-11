package model

import (
	"youtube-code-backend/pkg/common/types"
)

type MediaVariant struct {
	types.BaseModel
	AssetID  uint64 `gorm:"index;not null" json:"asset_id"`
	Quality  string `gorm:"size:20;not null" json:"quality"`
	URL      string `gorm:"size:500;not null" json:"url"`
	Width    int    `gorm:"default:0" json:"width"`
	Height   int    `gorm:"default:0" json:"height"`
	Bitrate  int64  `gorm:"default:0" json:"bitrate"`
	Codec    string `gorm:"size:50" json:"codec"`
	FileSize int64  `gorm:"default:0" json:"file_size"`
}

func (MediaVariant) TableName() string { return "media_variants" }
