package model

import "youtube-code-backend/pkg/common/types"

type VideoLike struct {
	types.BaseModel
	VideoID uint64 `gorm:"uniqueIndex:idx_video_likes_video_user;not null" json:"video_id"`
	UserID  uint64 `gorm:"uniqueIndex:idx_video_likes_video_user;not null" json:"user_id"`
	IsLike  bool   `gorm:"not null" json:"is_like"`
}

func (VideoLike) TableName() string { return "video_likes" }
