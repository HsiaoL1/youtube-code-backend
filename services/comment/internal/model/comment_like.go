package model

import "youtube-code-backend/pkg/common/types"

type CommentLike struct {
	types.BaseModel
	CommentID uint64 `gorm:"uniqueIndex:idx_comment_likes_comment_user;not null" json:"comment_id"`
	UserID    uint64 `gorm:"uniqueIndex:idx_comment_likes_comment_user;not null" json:"user_id"`
}

func (CommentLike) TableName() string { return "comment_likes" }
