package model

import "youtube-code-backend/pkg/common/types"

type NotificationType string

const (
	NotificationTypeNewVideo      NotificationType = "new_video"
	NotificationTypeNewLive       NotificationType = "new_live"
	NotificationTypeCommentReply  NotificationType = "comment_reply"
	NotificationTypeSubscription  NotificationType = "subscription"
	NotificationTypeLike          NotificationType = "like"
	NotificationTypeSystem        NotificationType = "system"
)

type Notification struct {
	types.BaseModel
	UserID       uint64           `gorm:"index;not null" json:"user_id"`
	Type         NotificationType `gorm:"size:50;not null" json:"type"`
	Title        string           `gorm:"size:255;not null" json:"title"`
	Body         string           `gorm:"type:text" json:"body"`
	IsRead       bool             `gorm:"default:false" json:"is_read"`
	ActorID      uint64           `gorm:"default:0" json:"actor_id"`
	ResourceType string           `gorm:"size:50" json:"resource_type"`
	ResourceID   uint64           `gorm:"default:0" json:"resource_id"`
}

func (Notification) TableName() string { return "notifications" }
