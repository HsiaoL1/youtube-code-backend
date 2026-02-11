package repository

import (
	"youtube-code-backend/services/chat/internal/model"

	"gorm.io/gorm"
)

// KeywordFilterRepository handles database operations for chat keyword filters.
type KeywordFilterRepository struct {
	db *gorm.DB
}

// NewKeywordFilterRepository creates a new KeywordFilterRepository.
func NewKeywordFilterRepository(db *gorm.DB) *KeywordFilterRepository {
	return &KeywordFilterRepository{db: db}
}

// Create inserts a new keyword filter.
func (r *KeywordFilterRepository) Create(filter *model.ChatKeywordFilter) error {
	return r.db.Create(filter).Error
}

// ListByRoomID returns all keyword filters for a given room.
func (r *KeywordFilterRepository) ListByRoomID(roomID uint64) ([]model.ChatKeywordFilter, error) {
	var filters []model.ChatKeywordFilter
	if err := r.db.Where("room_id = ?", roomID).Order("created_at ASC").Find(&filters).Error; err != nil {
		return nil, err
	}
	return filters, nil
}

// FindByID returns a keyword filter by its ID.
func (r *KeywordFilterRepository) FindByID(id uint64) (*model.ChatKeywordFilter, error) {
	var filter model.ChatKeywordFilter
	if err := r.db.First(&filter, id).Error; err != nil {
		return nil, err
	}
	return &filter, nil
}

// Delete removes a keyword filter by ID (hard delete).
func (r *KeywordFilterRepository) Delete(id uint64) error {
	result := r.db.Delete(&model.ChatKeywordFilter{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
