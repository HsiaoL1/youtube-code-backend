package repository

import (
	"youtube-code-backend/services/analytics/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LiveStatsRepository struct {
	db *gorm.DB
}

func NewLiveStatsRepository(db *gorm.DB) *LiveStatsRepository {
	return &LiveStatsRepository{db: db}
}

func (r *LiveStatsRepository) Upsert(stats *model.LiveSessionStats) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "session_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"peak_viewers", "avg_viewers", "total_messages",
			"duration_seconds", "unique_viewers", "updated_at",
		}),
	}).Create(stats).Error
}

func (r *LiveStatsRepository) FindBySessionID(sessionID uint64) (*model.LiveSessionStats, error) {
	var stats model.LiveSessionStats
	if err := r.db.Where("session_id = ?", sessionID).First(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}
