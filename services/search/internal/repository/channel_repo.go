package repository

import (
	"youtube-code-backend/services/search/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelRepository handles database operations for search channel entries.
type ChannelRepository struct {
	db *gorm.DB
}

// NewChannelRepository creates a new ChannelRepository.
func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

// Upsert inserts or updates a channel index entry by channel_id.
func (r *ChannelRepository) Upsert(ch *model.SearchChannel) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "channel_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"handle", "name", "description", "updated_at",
		}),
	}).Create(ch).Error
}

// Search performs a full-text search on channel name, handle, and description.
func (r *ChannelRepository) Search(query string, offset, limit int) ([]model.SearchChannel, int64, error) {
	var channels []model.SearchChannel
	var total int64

	baseQuery := r.db.Model(&model.SearchChannel{})

	if query != "" {
		baseQuery = baseQuery.Where(
			"to_tsvector('english', name || ' ' || coalesce(handle, '') || ' ' || coalesce(description, '')) @@ plainto_tsquery('english', ?)", query,
		)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := baseQuery.Order("created_at DESC").Offset(offset).Limit(limit).Find(&channels).Error; err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// SearchTop returns the top N channel results for a combined search.
func (r *ChannelRepository) SearchTop(query string, limit int) ([]model.SearchChannel, error) {
	var channels []model.SearchChannel

	baseQuery := r.db.Model(&model.SearchChannel{})

	if query != "" {
		baseQuery = baseQuery.Where(
			"to_tsvector('english', name || ' ' || coalesce(handle, '') || ' ' || coalesce(description, '')) @@ plainto_tsquery('english', ?)", query,
		)
	}

	if err := baseQuery.Order("created_at DESC").Limit(limit).Find(&channels).Error; err != nil {
		return nil, err
	}

	return channels, nil
}
