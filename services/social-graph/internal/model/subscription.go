package model

import (
	"youtube-code-backend/pkg/common/types"
)

// Subscription represents a user subscribing to a channel.
type Subscription struct {
	types.BaseModel
	SubscriberID     uint64 `gorm:"uniqueIndex:idx_subscriber_channel;not null" json:"subscriber_id"`
	ChannelID        uint64 `gorm:"uniqueIndex:idx_subscriber_channel;not null" json:"channel_id"`
	NotifyPreference string `gorm:"size:50;default:all;not null" json:"notify_preference"`
}

func (Subscription) TableName() string { return "subscriptions" }
