package model

import "youtube-code-backend/pkg/common/types"

type Playlist struct {
	types.BaseModel
	UserID      uint64         `gorm:"index;not null" json:"user_id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Visibility  string         `gorm:"size:20;default:private;not null" json:"visibility"`
	VideoCount  int64          `gorm:"default:0" json:"video_count"`
	Items       []PlaylistItem `gorm:"foreignKey:PlaylistID" json:"items,omitempty"`
}

func (Playlist) TableName() string { return "playlists" }

type PlaylistItem struct {
	types.BaseModel
	PlaylistID uint64 `gorm:"index;not null" json:"playlist_id"`
	VideoID    uint64 `gorm:"not null" json:"video_id"`
	Position   int    `gorm:"not null" json:"position"`
}

func (PlaylistItem) TableName() string { return "playlist_items" }
