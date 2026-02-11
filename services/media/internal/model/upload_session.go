package model

import (
	"youtube-code-backend/pkg/common/types"
)

type UploadStatus string

const (
	UploadStatusInitiated UploadStatus = "initiated"
	UploadStatusUploading UploadStatus = "uploading"
	UploadStatusCompleted UploadStatus = "completed"
	UploadStatusFailed    UploadStatus = "failed"
)

type UploadSession struct {
	types.BaseModel
	SessionUUID   string       `gorm:"uniqueIndex;size:36;not null" json:"session_uuid"`
	UserID        uint64       `gorm:"index;not null" json:"user_id"`
	Filename      string       `gorm:"size:255;not null" json:"filename"`
	FileSize      int64        `gorm:"not null" json:"file_size"`
	MimeType      string       `gorm:"size:100;not null" json:"mime_type"`
	Status        UploadStatus `gorm:"size:20;default:initiated;not null" json:"status"`
	StorageKey    string       `gorm:"size:500" json:"storage_key"`
	PartsTotal    int          `gorm:"default:0" json:"parts_total"`
	PartsUploaded int          `gorm:"default:0" json:"parts_uploaded"`
}

func (UploadSession) TableName() string { return "upload_sessions" }
