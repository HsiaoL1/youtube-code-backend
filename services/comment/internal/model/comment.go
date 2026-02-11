package model

import "youtube-code-backend/pkg/common/types"

type CommentStatus string

const (
	CommentStatusActive  CommentStatus = "active"
	CommentStatusDeleted CommentStatus = "deleted"
	CommentStatusHidden  CommentStatus = "hidden"
)

type Comment struct {
	types.BaseModel
	VideoID    uint64        `gorm:"index;not null" json:"video_id"`
	UserID     uint64        `gorm:"index;not null" json:"user_id"`
	ParentID   *uint64       `gorm:"index" json:"parent_id,omitempty"`
	Content    string        `gorm:"type:text;not null" json:"content"`
	LikeCount  int64         `gorm:"default:0" json:"like_count"`
	ReplyCount int64         `gorm:"default:0" json:"reply_count"`
	IsPinned   bool          `gorm:"default:false" json:"is_pinned"`
	IsHearted  bool          `gorm:"default:false" json:"is_hearted"`
	Status     CommentStatus `gorm:"size:20;default:active;not null" json:"status"`
}

func (Comment) TableName() string { return "comments" }
