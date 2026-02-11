package repository

import (
	"youtube-code-backend/services/analytics/internal/model"

	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *model.AnalyticsEvent) error {
	return r.db.Create(event).Error
}

func (r *EventRepository) FindByVideoID(videoID uint64, offset, limit int) ([]model.AnalyticsEvent, int64, error) {
	var events []model.AnalyticsEvent
	var total int64

	query := r.db.Model(&model.AnalyticsEvent{}).Where("video_id = ?", videoID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *EventRepository) FindByChannelID(channelID uint64, offset, limit int) ([]model.AnalyticsEvent, int64, error) {
	var events []model.AnalyticsEvent
	var total int64

	query := r.db.Model(&model.AnalyticsEvent{}).Where("channel_id = ?", channelID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

// CountByVideoAndDate returns events within a date range. If videoID is 0,
// it returns all events with a non-zero video_id (used for aggregation discovery).
func (r *EventRepository) CountByVideoAndDate(videoID uint64, startDate, endDate string) ([]model.AnalyticsEvent, error) {
	var events []model.AnalyticsEvent
	query := r.db.Where("DATE(created_at) BETWEEN ? AND ?", startDate, endDate)
	if videoID > 0 {
		query = query.Where("video_id = ?", videoID)
	} else {
		query = query.Where("video_id > 0")
	}
	err := query.Order("created_at DESC").Find(&events).Error
	return events, err
}

// CountByChannelAndDate returns events within a date range. If channelID is 0,
// it returns all events with a non-zero channel_id (used for aggregation discovery).
func (r *EventRepository) CountByChannelAndDate(channelID uint64, startDate, endDate string) ([]model.AnalyticsEvent, error) {
	var events []model.AnalyticsEvent
	query := r.db.Where("DATE(created_at) BETWEEN ? AND ?", startDate, endDate)
	if channelID > 0 {
		query = query.Where("channel_id = ?", channelID)
	} else {
		query = query.Where("channel_id > 0")
	}
	err := query.Order("created_at DESC").Find(&events).Error
	return events, err
}

// CountRecentByVideo counts events for a video in the last N minutes.
func (r *EventRepository) CountRecentByVideo(videoID uint64, minutes int) (int64, error) {
	var count int64
	err := r.db.Model(&model.AnalyticsEvent{}).
		Where("video_id = ? AND created_at >= NOW() - INTERVAL '1 minute' * ?", videoID, minutes).
		Count(&count).Error
	return count, err
}
