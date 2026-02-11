package model

import (
	"youtube-code-backend/pkg/common/types"
)

type EventType string

const (
	EventTypeView      EventType = "view"
	EventTypeWatchTime EventType = "watch_time"
	EventTypeLike      EventType = "like"
	EventTypeShare     EventType = "share"
	EventTypeSearch    EventType = "search"
	EventTypeClick     EventType = "click"
)

// ValidEventTypes contains all valid event type values.
var ValidEventTypes = map[EventType]bool{
	EventTypeView:      true,
	EventTypeWatchTime: true,
	EventTypeLike:      true,
	EventTypeShare:     true,
	EventTypeSearch:    true,
	EventTypeClick:     true,
}

type AnalyticsEvent struct {
	types.BaseModel
	EventType  string `gorm:"not null;index" json:"event_type"`
	VideoID    uint64 `gorm:"index" json:"video_id"`
	ChannelID  uint64 `gorm:"index" json:"channel_id"`
	UserID     uint64 `json:"user_id"`
	SessionID  string `json:"session_id"`
	Properties string `json:"properties"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	Country    string `json:"country"`
	DeviceType string `json:"device_type"`
}

func (AnalyticsEvent) TableName() string { return "analytics_events" }
