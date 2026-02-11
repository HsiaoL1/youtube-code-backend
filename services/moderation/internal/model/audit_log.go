package model

import "youtube-code-backend/pkg/common/types"

type AuditLog struct {
	types.BaseModel
	ActorID    uint64 `gorm:"index" json:"actor_id"`
	Action     string `gorm:"not null" json:"action"`
	TargetType string `json:"target_type"`
	TargetID   uint64 `json:"target_id"`
	Details    string `gorm:"type:text" json:"details"`
	IPAddress  string `json:"ip_address"`
}

func (AuditLog) TableName() string { return "audit_logs" }
