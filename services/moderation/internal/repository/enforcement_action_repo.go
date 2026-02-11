package repository

import (
	"youtube-code-backend/services/moderation/internal/model"

	"gorm.io/gorm"
)

type EnforcementActionRepository struct {
	db *gorm.DB
}

func NewEnforcementActionRepository(db *gorm.DB) *EnforcementActionRepository {
	return &EnforcementActionRepository{db: db}
}

func (r *EnforcementActionRepository) Create(action *model.EnforcementAction) error {
	return r.db.Create(action).Error
}

func (r *EnforcementActionRepository) FindByUserID(userID uint64, offset, limit int) ([]model.EnforcementAction, int64, error) {
	var actions []model.EnforcementAction
	var total int64

	query := r.db.Model(&model.EnforcementAction{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&actions).Error; err != nil {
		return nil, 0, err
	}
	return actions, total, nil
}
