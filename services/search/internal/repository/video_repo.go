package repository

import (
	"youtube-code-backend/services/search/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// VideoRepository handles database operations for search video entries.
type VideoRepository struct {
	db *gorm.DB
}

// NewVideoRepository creates a new VideoRepository.
func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

// Upsert inserts or updates a video index entry by video_id.
func (r *VideoRepository) Upsert(v *model.SearchVideo) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "video_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title", "description", "channel_name", "category", "tags", "visibility", "status", "updated_at",
		}),
	}).Create(v).Error
}

// DeleteByVideoID removes a video from the search index.
func (r *VideoRepository) DeleteByVideoID(videoID uint64) error {
	result := r.db.Where("video_id = ?", videoID).Delete(&model.SearchVideo{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Search performs a full-text search on video title and description.
func (r *VideoRepository) Search(query, category, sort string, offset, limit int) ([]model.SearchVideo, int64, error) {
	var videos []model.SearchVideo
	var total int64

	baseQuery := r.db.Model(&model.SearchVideo{}).
		Where("visibility = ? AND status = ?", "public", "ready")

	if query != "" {
		baseQuery = baseQuery.Where(
			"to_tsvector('english', title || ' ' || coalesce(description, '')) @@ plainto_tsquery('english', ?)", query,
		)
	}

	if category != "" {
		baseQuery = baseQuery.Where("category = ?", category)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if query != "" && (sort == "relevance" || sort == "") {
		baseQuery = baseQuery.Order(
			gorm.Expr("ts_rank_cd(to_tsvector('english', title || ' ' || coalesce(description, '')), plainto_tsquery('english', ?)) DESC", query),
		)
	} else {
		baseQuery = baseQuery.Order("created_at DESC")
	}

	if err := baseQuery.Offset(offset).Limit(limit).Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// SearchTop returns the top N video results for a combined search.
func (r *VideoRepository) SearchTop(query string, limit int) ([]model.SearchVideo, error) {
	var videos []model.SearchVideo

	baseQuery := r.db.Model(&model.SearchVideo{}).
		Where("visibility = ? AND status = ?", "public", "ready")

	if query != "" {
		baseQuery = baseQuery.Where(
			"to_tsvector('english', title || ' ' || coalesce(description, '')) @@ plainto_tsquery('english', ?)", query,
		)
	}

	if err := baseQuery.Order("created_at DESC").Limit(limit).Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}
