package model

import (
	"time"

	"youtube-code-backend/pkg/common/types"
)

// VideoType mirrors the video service's VideoType enum.
type VideoType string

const (
	VideoTypeVideo VideoType = "video"
	VideoTypeShort VideoType = "short"
)

// VideoStatus mirrors the video service's VideoStatus enum.
type VideoStatus string

const (
	VideoStatusReady VideoStatus = "ready"
)

// VideoVisibility mirrors the video service's VideoVisibility enum.
type VideoVisibility string

const (
	VideoVisibilityPublic VideoVisibility = "public"
)

// Video is a read-only reference to the videos table used for feed queries.
type Video struct {
	types.BaseModel
	ChannelID     uint64          `gorm:"index;not null" json:"channel_id"`
	Type          VideoType       `gorm:"size:20;not null" json:"type"`
	Title         string          `gorm:"size:255;not null" json:"title"`
	Description   string          `gorm:"type:text" json:"description"`
	Status        VideoStatus     `gorm:"size:20;not null" json:"status"`
	Visibility    VideoVisibility `gorm:"size:20;not null" json:"visibility"`
	Duration      int64           `gorm:"default:0" json:"duration"`
	ViewCount     int64           `gorm:"default:0" json:"view_count"`
	LikeCount     int64           `gorm:"default:0" json:"like_count"`
	DislikeCount  int64           `gorm:"default:0" json:"dislike_count"`
	CommentCount  int64           `gorm:"default:0" json:"comment_count"`
	FavoriteCount int64           `gorm:"default:0" json:"favorite_count"`
	ThumbnailURL  string          `gorm:"size:500" json:"thumbnail_url"`
	ScheduledAt   *time.Time      `json:"scheduled_at,omitempty"`
	Category      string          `gorm:"-" json:"category,omitempty"`
}

func (Video) TableName() string { return "videos" }

// Subscription is a read-only reference to the subscriptions table.
type Subscription struct {
	types.BaseModel
	SubscriberID     uint64 `gorm:"uniqueIndex:idx_subscriber_channel;not null" json:"subscriber_id"`
	ChannelID        uint64 `gorm:"uniqueIndex:idx_subscriber_channel;not null" json:"channel_id"`
	NotifyPreference string `gorm:"size:50;default:all;not null" json:"notify_preference"`
}

func (Subscription) TableName() string { return "subscriptions" }
