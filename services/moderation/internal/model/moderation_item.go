package model

import (
	"time"

	"youtube-code-backend/pkg/common/types"
)

type ModerationStatus string

const (
	ModerationStatusPending  ModerationStatus = "pending"
	ModerationStatusApproved ModerationStatus = "approved"
	ModerationStatusRejected ModerationStatus = "rejected"
)

type ContentType string

const (
	ContentTypeVideo    ContentType = "video"
	ContentTypeComment  ContentType = "comment"
	ContentTypeLiveRoom ContentType = "live_room"
	ContentTypeChannel  ContentType = "channel"
)

type ModerationItem struct {
	types.BaseModel
	ContentType     string     `gorm:"not null;index" json:"content_type"`
	ContentID       uint64     `gorm:"not null" json:"content_id"`
	ReportedBy      uint64     `json:"reported_by"`
	Status          string     `gorm:"default:pending;index" json:"status"`
	Decision        string     `json:"decision"`
	RejectionReason string     `gorm:"type:text" json:"rejection_reason"`
	ReviewedBy      uint64     `json:"reviewed_by"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
}

func (ModerationItem) TableName() string { return "moderation_items" }

// ValidContentTypes is the set of allowed content types.
var ValidContentTypes = map[string]bool{
	string(ContentTypeVideo):    true,
	string(ContentTypeComment):  true,
	string(ContentTypeLiveRoom): true,
	string(ContentTypeChannel):  true,
}
