package model

import (
	"youtube-code-backend/pkg/common/types"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
)

type UserRole string

const (
	UserRoleUser      UserRole = "user"
	UserRoleModerator UserRole = "moderator"
	UserRoleAdmin     UserRole = "admin"
)

type User struct {
	types.BaseModel
	Username     string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone        string     `gorm:"size:20" json:"phone,omitempty"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Role         UserRole   `gorm:"size:20;default:user;not null" json:"role"`
	Status       UserStatus `gorm:"size:20;default:active;not null" json:"status"`
}

func (User) TableName() string { return "users" }

type RefreshToken struct {
	types.BaseModel
	TokenHash string `gorm:"uniqueIndex;size:255;not null"`
	UserID    uint64 `gorm:"index;not null"`
	ExpiresAt int64  `gorm:"not null"`
	Revoked   bool   `gorm:"default:false;not null"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type VerificationCode struct {
	types.BaseModel
	UserID    uint64 `gorm:"index;not null"`
	Code      string `gorm:"size:10;not null"`
	Type      string `gorm:"size:20;not null"` // email, phone
	ExpiresAt int64  `gorm:"not null"`
	Used      bool   `gorm:"default:false;not null"`
}

func (VerificationCode) TableName() string { return "verification_codes" }
