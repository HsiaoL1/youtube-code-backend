package repository

import (
	"time"

	"youtube-code-backend/services/analytics/internal/model"

	"gorm.io/gorm"
)

type CounterRepository struct {
	db *gorm.DB
}

func NewCounterRepository(db *gorm.DB) *CounterRepository {
	return &CounterRepository{db: db}
}

func (r *CounterRepository) Create(snapshot *model.CounterSnapshot) error {
	return r.db.Create(snapshot).Error
}

// FindLatest returns the most recent snapshot for a given entity and counter.
func (r *CounterRepository) FindLatest(entityType string, entityID uint64, counterName string) (*model.CounterSnapshot, error) {
	var snapshot model.CounterSnapshot
	err := r.db.Where("entity_type = ? AND entity_id = ? AND counter_name = ?", entityType, entityID, counterName).
		Order("snapshotted_at DESC").
		First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

// Increment finds the latest snapshot for the counter and creates a new snapshot
// with the value incremented by delta. If no snapshot exists, creates one with value = delta.
func (r *CounterRepository) Increment(entityType string, entityID uint64, counterName string, delta int64) (*model.CounterSnapshot, error) {
	var currentValue int64

	existing, err := r.FindLatest(entityType, entityID, counterName)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if existing != nil {
		currentValue = existing.Value
	}

	snapshot := &model.CounterSnapshot{
		EntityType:    entityType,
		EntityID:      entityID,
		CounterName:   counterName,
		Value:         currentValue + delta,
		SnapshottedAt: time.Now(),
	}

	if err := r.db.Create(snapshot).Error; err != nil {
		return nil, err
	}
	return snapshot, nil
}
