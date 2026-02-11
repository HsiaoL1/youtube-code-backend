package model

import "youtube-code-backend/pkg/common/types"

type VideoFavorite struct {
	types.BaseModel
	VideoID uint64 `gorm:"uniqueIndex:idx_video_favorites_video_user;not null" json:"video_id"`
	UserID  uint64 `gorm:"uniqueIndex:idx_video_favorites_video_user;not null" json:"user_id"`
}

func (VideoFavorite) TableName() string { return "video_favorites" }
