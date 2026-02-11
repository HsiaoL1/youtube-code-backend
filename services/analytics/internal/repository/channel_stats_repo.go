package repository

import (
	"youtube-code-backend/services/analytics/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChannelStatsRepository struct {
	db *gorm.DB
}

func NewChannelStatsRepository(db *gorm.DB) *ChannelStatsRepository {
	return &ChannelStatsRepository{db: db}
}

func (r *ChannelStatsRepository) Upsert(stats *model.ChannelDailyStats) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "channel_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"views", "watch_time_seconds", "new_subscribers",
			"lost_subscribers", "revenue_cents", "updated_at",
		}),
	}).Create(stats).Error
}

func (r *ChannelStatsRepository) FindByChannelID(channelID uint64, startDate, endDate string) ([]model.ChannelDailyStats, error) {
	var stats []model.ChannelDailyStats
	query := r.db.Where("channel_id = ?", channelID)
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}
	err := query.Order("date ASC").Find(&stats).Error
	return stats, err
}

// SumByChannelID returns the total aggregated stats for a channel across all dates.
func (r *ChannelStatsRepository) SumByChannelID(channelID uint64) (*model.ChannelDailyStats, error) {
	var result model.ChannelDailyStats
	err := r.db.Model(&model.ChannelDailyStats{}).
		Select(`
			? as channel_id,
			COALESCE(SUM(views), 0) as views,
			COALESCE(SUM(watch_time_seconds), 0) as watch_time_seconds,
			COALESCE(SUM(new_subscribers), 0) as new_subscribers,
			COALESCE(SUM(lost_subscribers), 0) as lost_subscribers,
			COALESCE(SUM(revenue_cents), 0) as revenue_cents
		`, channelID).
		Where("channel_id = ?", channelID).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	result.ChannelID = channelID
	return &result, nil
}
