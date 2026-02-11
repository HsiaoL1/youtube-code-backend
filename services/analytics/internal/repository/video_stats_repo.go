package repository

import (
	"youtube-code-backend/services/analytics/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VideoStatsRepository struct {
	db *gorm.DB
}

func NewVideoStatsRepository(db *gorm.DB) *VideoStatsRepository {
	return &VideoStatsRepository{db: db}
}

func (r *VideoStatsRepository) Upsert(stats *model.VideoDailyStats) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "video_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"views", "watch_time_seconds", "likes", "dislikes",
			"shares", "comments", "unique_viewers", "updated_at",
		}),
	}).Create(stats).Error
}

func (r *VideoStatsRepository) FindByVideoID(videoID uint64, startDate, endDate string) ([]model.VideoDailyStats, error) {
	var stats []model.VideoDailyStats
	query := r.db.Where("video_id = ?", videoID)
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}
	err := query.Order("date ASC").Find(&stats).Error
	return stats, err
}

// SumByVideoID returns the total aggregated stats for a video across all dates.
func (r *VideoStatsRepository) SumByVideoID(videoID uint64) (*model.VideoDailyStats, error) {
	var result model.VideoDailyStats
	err := r.db.Model(&model.VideoDailyStats{}).
		Select(`
			? as video_id,
			COALESCE(SUM(views), 0) as views,
			COALESCE(SUM(watch_time_seconds), 0) as watch_time_seconds,
			COALESCE(SUM(likes), 0) as likes,
			COALESCE(SUM(dislikes), 0) as dislikes,
			COALESCE(SUM(shares), 0) as shares,
			COALESCE(SUM(comments), 0) as comments,
			COALESCE(SUM(unique_viewers), 0) as unique_viewers
		`, videoID).
		Where("video_id = ?", videoID).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	result.VideoID = videoID
	return &result, nil
}

// IncrementColumn atomically increments a column for a video on a given date.
func (r *VideoStatsRepository) IncrementColumn(videoID uint64, date string, column string, delta int64) error {
	result := r.db.Model(&model.VideoDailyStats{}).
		Where("video_id = ? AND date = ?", videoID, date).
		Update(column, gorm.Expr(column+" + ?", delta))

	if result.Error != nil {
		return result.Error
	}

	// If no row was updated, create one.
	if result.RowsAffected == 0 {
		stats := &model.VideoDailyStats{
			VideoID: videoID,
			Date:    date,
		}
		// Set the specific column via a map so we can handle dynamic column names.
		if err := r.db.Create(stats).Error; err != nil {
			return err
		}
		return r.db.Model(stats).Update(column, gorm.Expr(column+" + ?", delta)).Error
	}
	return nil
}
