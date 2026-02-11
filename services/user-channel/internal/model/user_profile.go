package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"youtube-code-backend/pkg/common/types"
)

// StringSlice is a custom type for storing JSON string arrays in a text column.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshal StringSlice: %w", err)
	}
	return string(b), nil
}

func (s *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*s = StringSlice{}
		return nil
	}
	var bytes []byte
	switch v := src.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("unsupported type for StringSlice: %T", src)
	}
	return json.Unmarshal(bytes, s)
}

type UserProfile struct {
	types.BaseModel
	UserID        uint64      `gorm:"uniqueIndex;not null" json:"user_id"`
	Nickname      string      `gorm:"size:100" json:"nickname"`
	Avatar        string      `gorm:"size:500" json:"avatar"`
	Bio           string      `gorm:"size:1000" json:"bio"`
	Region        string      `gorm:"size:50" json:"region"`
	Gender        string      `gorm:"size:20" json:"gender"`
	Links         StringSlice `gorm:"type:text" json:"links"`
	AccountStatus string      `gorm:"size:20;default:active;not null" json:"account_status"`
}

func (UserProfile) TableName() string { return "user_profiles" }
