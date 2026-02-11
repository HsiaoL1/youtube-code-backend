package repository

import (
	"youtube-code-backend/services/chat/internal/model"

	"gorm.io/gorm"
)

// ManagerRepository handles database operations for chat managers.
type ManagerRepository struct {
	db *gorm.DB
}

// NewManagerRepository creates a new ManagerRepository.
func NewManagerRepository(db *gorm.DB) *ManagerRepository {
	return &ManagerRepository{db: db}
}

// Create inserts a new chat manager record.
func (r *ManagerRepository) Create(mgr *model.ChatManager) error {
	return r.db.Create(mgr).Error
}

// FindByRoomAndUser finds a manager by room ID and user ID.
func (r *ManagerRepository) FindByRoomAndUser(roomID, userID uint64) (*model.ChatManager, error) {
	var mgr model.ChatManager
	if err := r.db.Where("room_id = ? AND user_id = ?", roomID, userID).First(&mgr).Error; err != nil {
		return nil, err
	}
	return &mgr, nil
}

// ListByRoomID returns all managers for a given room.
func (r *ManagerRepository) ListByRoomID(roomID uint64) ([]model.ChatManager, error) {
	var managers []model.ChatManager
	if err := r.db.Where("room_id = ?", roomID).Order("created_at ASC").Find(&managers).Error; err != nil {
		return nil, err
	}
	return managers, nil
}

// Delete removes a manager by room ID and user ID (hard delete).
func (r *ManagerRepository) Delete(roomID, userID uint64) error {
	result := r.db.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&model.ChatManager{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
