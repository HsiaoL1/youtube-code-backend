package model

import "youtube-code-backend/pkg/common/types"

type SensitiveWord struct {
	types.BaseModel
	Word     string `gorm:"uniqueIndex;size:200;not null" json:"word"`
	Severity string `gorm:"size:20;default:warn;not null" json:"severity"`
	IsActive bool   `gorm:"default:true;not null" json:"is_active"`
}

func (SensitiveWord) TableName() string { return "sensitive_words" }
