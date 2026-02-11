package repository

import (
	"youtube-code-backend/services/comment/internal/model"

	"gorm.io/gorm"
)

type SensitiveWordRepository struct {
	db *gorm.DB
}

func NewSensitiveWordRepository(db *gorm.DB) *SensitiveWordRepository {
	return &SensitiveWordRepository{db: db}
}

func (r *SensitiveWordRepository) FindAll() ([]model.SensitiveWord, error) {
	var words []model.SensitiveWord
	if err := r.db.Order("created_at DESC").Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
}

func (r *SensitiveWordRepository) FindActive() ([]model.SensitiveWord, error) {
	var words []model.SensitiveWord
	if err := r.db.Where("is_active = ?", true).Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
}

func (r *SensitiveWordRepository) FindByWord(word string) (*model.SensitiveWord, error) {
	var sw model.SensitiveWord
	if err := r.db.Where("word = ?", word).First(&sw).Error; err != nil {
		return nil, err
	}
	return &sw, nil
}

func (r *SensitiveWordRepository) Create(sw *model.SensitiveWord) error {
	return r.db.Create(sw).Error
}

func (r *SensitiveWordRepository) Delete(id uint64) error {
	return r.db.Unscoped().Delete(&model.SensitiveWord{}, id).Error
}
