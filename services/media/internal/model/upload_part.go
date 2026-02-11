package model

import (
	"youtube-code-backend/pkg/common/types"
)

type PartStatus string

const (
	PartStatusPending  PartStatus = "pending"
	PartStatusUploaded PartStatus = "uploaded"
	PartStatusFailed   PartStatus = "failed"
)

type UploadPart struct {
	types.BaseModel
	SessionID  uint64     `gorm:"index;not null" json:"session_id"`
	PartNumber int        `gorm:"not null" json:"part_number"`
	Size       int64      `gorm:"not null" json:"size"`
	ETag       string     `gorm:"size:255" json:"etag"`
	Status     PartStatus `gorm:"size:20;default:pending;not null" json:"status"`
}

func (UploadPart) TableName() string { return "upload_parts" }
