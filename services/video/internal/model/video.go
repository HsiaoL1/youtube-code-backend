package model

import (
	"time"

	"youtube-code-backend/pkg/common/types"
)

type VideoType string

const (
	VideoTypeVideo VideoType = "video"
	VideoTypeShort VideoType = "short"
)

type VideoStatus string

const (
	VideoStatusDraft      VideoStatus = "draft"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusReady      VideoStatus = "ready"
	VideoStatusRemoved    VideoStatus = "removed"
	VideoStatusRejected   VideoStatus = "rejected"
)

type VideoVisibility string

const (
	VideoVisibilityPrivate  VideoVisibility = "private"
	VideoVisibilityPublic   VideoVisibility = "public"
	VideoVisibilityUnlisted VideoVisibility = "unlisted"
)

type Video struct {
	types.BaseModel
	ChannelID     uint64          `gorm:"index;not null" json:"channel_id"`
	Type          VideoType       `gorm:"size:20;default:video;not null" json:"type"`
	Title         string          `gorm:"size:255;not null" json:"title"`
	Description   string          `gorm:"type:text" json:"description"`
	Status        VideoStatus     `gorm:"size:20;default:draft;not null" json:"status"`
	Visibility    VideoVisibility `gorm:"size:20;default:private;not null" json:"visibility"`
	Duration      int64           `gorm:"default:0" json:"duration"`
	ViewCount     int64           `gorm:"default:0" json:"view_count"`
	LikeCount     int64           `gorm:"default:0" json:"like_count"`
	DislikeCount  int64           `gorm:"default:0" json:"dislike_count"`
	CommentCount  int64           `gorm:"default:0" json:"comment_count"`
	FavoriteCount int64           `gorm:"default:0" json:"favorite_count"`
	ThumbnailURL  string          `gorm:"size:500" json:"thumbnail_url"`
	ScheduledAt   *time.Time      `json:"scheduled_at,omitempty"`
}

func (Video) TableName() string { return "videos" }

// ValidStatusTransitions defines which status transitions are allowed.
var ValidStatusTransitions = map[VideoStatus][]VideoStatus{
	VideoStatusDraft:      {VideoStatusProcessing},
	VideoStatusProcessing: {VideoStatusReady},
	VideoStatusReady:      {VideoStatusRemoved, VideoStatusRejected},
}

// CanTransitionTo checks whether the video can move to the target status.
func (v *Video) CanTransitionTo(target VideoStatus) bool {
	allowed, ok := ValidStatusTransitions[v.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}
