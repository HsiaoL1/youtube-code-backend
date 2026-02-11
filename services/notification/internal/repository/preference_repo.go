package repository

import (
	"youtube-code-backend/services/notification/internal/model"

	"gorm.io/gorm"
)

type PreferenceRepository struct {
	db *gorm.DB
}

func NewPreferenceRepository(db *gorm.DB) *PreferenceRepository {
	return &PreferenceRepository{db: db}
}

func (r *PreferenceRepository) FindByUserID(userID uint64) (*model.NotificationPreference, error) {
	var pref model.NotificationPreference
	if err := r.db.Where("user_id = ?", userID).First(&pref).Error; err != nil {
		return nil, err
	}
	return &pref, nil
}

func (r *PreferenceRepository) Upsert(pref *model.NotificationPreference) error {
	return r.db.Save(pref).Error
}
