package model

import (
	"youtube-code-backend/pkg/common/types"
)

type RoomStatus string

const (
	RoomStatusIdle  RoomStatus = "idle"
	RoomStatusLive  RoomStatus = "live"
	RoomStatusEnded RoomStatus = "ended"
)

// ValidRoomStatusTransitions defines which status transitions are allowed.
var ValidRoomStatusTransitions = map[RoomStatus][]RoomStatus{
	RoomStatusIdle: {RoomStatusLive},
	RoomStatusLive: {RoomStatusEnded},
	RoomStatusEnded: {RoomStatusIdle},
}

// CanTransitionTo checks whether the room can move to the target status.
func (r *LiveRoom) CanTransitionTo(target RoomStatus) bool {
	allowed, ok := ValidRoomStatusTransitions[r.Status]
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

type LiveRoom struct {
	types.BaseModel
	ChannelID       uint64     `gorm:"index;not null" json:"channel_id"`
	Title           string     `gorm:"size:255;not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	Status          RoomStatus `gorm:"size:20;default:idle;not null" json:"status"`
	StreamKey       string     `gorm:"uniqueIndex;size:255;not null" json:"-"`
	PublishURL      string     `gorm:"size:500" json:"publish_url,omitempty"`
	PlaybackURL     string     `gorm:"size:500" json:"playback_url,omitempty"`
	ViewerCount     int64      `gorm:"default:0" json:"viewer_count"`
	PeakViewerCount int64      `gorm:"default:0" json:"peak_viewer_count"`
	Category        string     `gorm:"size:100" json:"category"`
	ThumbnailURL    string     `gorm:"size:500" json:"thumbnail_url"`
}

func (LiveRoom) TableName() string { return "live_rooms" }
