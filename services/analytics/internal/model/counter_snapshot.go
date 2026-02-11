package model

import (
	"time"

	"youtube-code-backend/pkg/common/types"
)

type CounterSnapshot struct {
	types.BaseModel
	EntityType    string    `gorm:"index:idx_entity_counter" json:"entity_type"`
	EntityID      uint64    `gorm:"index:idx_entity_counter" json:"entity_id"`
	CounterName   string    `gorm:"index:idx_entity_counter" json:"counter_name"`
	Value         int64     `gorm:"default:0" json:"value"`
	SnapshottedAt time.Time `json:"snapshotted_at"`
}

func (CounterSnapshot) TableName() string { return "counter_snapshots" }
