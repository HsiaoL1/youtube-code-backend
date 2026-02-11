package model

import "youtube-code-backend/pkg/common/types"

type NotificationPreference struct {
	types.BaseModel
	UserID              uint64 `gorm:"uniqueIndex;not null" json:"user_id"`
	NewVideo            bool   `gorm:"default:true" json:"new_video"`
	NewLive             bool   `gorm:"default:true" json:"new_live"`
	CommentReply        bool   `gorm:"default:true" json:"comment_reply"`
	Subscription        bool   `gorm:"default:true" json:"subscription"`
	Likes               bool   `gorm:"default:true" json:"likes"`
	SystemNotifications bool   `gorm:"default:true" json:"system_notifications"`
}

func (NotificationPreference) TableName() string { return "notification_preferences" }
