package repository

import (
	"youtube-code-backend/services/search/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LiveRepository handles database operations for search live room entries.
type LiveRepository struct {
	db *gorm.DB
}

// NewLiveRepository creates a new LiveRepository.
func NewLiveRepository(db *gorm.DB) *LiveRepository {
	return &LiveRepository{db: db}
}

// Upsert inserts or updates a live room index entry by room_id.
func (r *LiveRepository) Upsert(live *model.SearchLive) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "room_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title", "channel_name", "category", "is_live", "updated_at",
		}),
	}).Create(live).Error
}

// Search performs a full-text search on live room title and channel name.
func (r *LiveRepository) Search(query string, offset, limit int) ([]model.SearchLive, int64, error) {
	var rooms []model.SearchLive
	var total int64

	baseQuery := r.db.Model(&model.SearchLive{})

	if query != "" {
		baseQuery = baseQuery.Where(
			"to_tsvector('english', title || ' ' || coalesce(channel_name, '')) @@ plainto_tsquery('english', ?)", query,
		)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := baseQuery.Order("is_live DESC, created_at DESC").Offset(offset).Limit(limit).Find(&rooms).Error; err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

// SearchTop returns the top N live room results for a combined search.
func (r *LiveRepository) SearchTop(query string, limit int) ([]model.SearchLive, error) {
	var rooms []model.SearchLive

	baseQuery := r.db.Model(&model.SearchLive{})

	if query != "" {
		baseQuery = baseQuery.Where(
			"to_tsvector('english', title || ' ' || coalesce(channel_name, '')) @@ plainto_tsquery('english', ?)", query,
		)
	}

	if err := baseQuery.Order("is_live DESC, created_at DESC").Limit(limit).Find(&rooms).Error; err != nil {
		return nil, err
	}

	return rooms, nil
}
