package repository

import (
	"youtube-code-backend/services/user-channel/internal/model"

	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) FindByUserID(userID uint64) (*model.UserProfile, error) {
	var profile model.UserProfile
	if err := r.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ProfileRepository) Upsert(profile *model.UserProfile) error {
	// Try to find existing profile
	var existing model.UserProfile
	err := r.db.Where("user_id = ?", profile.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(profile).Error
	}
	if err != nil {
		return err
	}

	// Update existing
	profile.ID = existing.ID
	return r.db.Save(profile).Error
}
