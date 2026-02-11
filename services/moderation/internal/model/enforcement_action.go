package model

import (
	"time"

	"youtube-code-backend/pkg/common/types"
)

type ActionType string

const (
	ActionTypeWarning ActionType = "warning"
	ActionTypeMute    ActionType = "mute"
	ActionTypeSuspend ActionType = "suspend"
	ActionTypeBan     ActionType = "ban"
)

type EnforcementAction struct {
	types.BaseModel
	UserID      uint64     `gorm:"index" json:"user_id"`
	ActionType  string     `gorm:"not null" json:"action_type"`
	Reason      string     `gorm:"type:text" json:"reason"`
	PerformedBy uint64     `json:"performed_by"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

func (EnforcementAction) TableName() string { return "enforcement_actions" }

// ValidActionTypes is the set of allowed enforcement action types.
var ValidActionTypes = map[string]bool{
	string(ActionTypeWarning): true,
	string(ActionTypeMute):    true,
	string(ActionTypeSuspend): true,
	string(ActionTypeBan):     true,
}
