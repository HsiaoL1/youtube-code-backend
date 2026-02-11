package repository

import (
	"youtube-code-backend/services/moderation/internal/model"

	"gorm.io/gorm"
)

type ModerationItemRepository struct {
	db *gorm.DB
}

func NewModerationItemRepository(db *gorm.DB) *ModerationItemRepository {
	return &ModerationItemRepository{db: db}
}

func (r *ModerationItemRepository) Create(item *model.ModerationItem) error {
	return r.db.Create(item).Error
}

func (r *ModerationItemRepository) FindByID(id uint64) (*model.ModerationItem, error) {
	var item model.ModerationItem
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ModerationItemRepository) Update(item *model.ModerationItem) error {
	return r.db.Save(item).Error
}

func (r *ModerationItemRepository) FindPending(contentType string, offset, limit int) ([]model.ModerationItem, int64, error) {
	var items []model.ModerationItem
	var total int64

	query := r.db.Model(&model.ModerationItem{}).Where("status = ?", string(model.ModerationStatusPending))
	if contentType != "" {
		query = query.Where("content_type = ?", contentType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at ASC").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
