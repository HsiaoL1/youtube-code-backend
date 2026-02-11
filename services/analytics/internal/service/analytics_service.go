package service

import (
	"time"

	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/services/analytics/internal/model"
	"youtube-code-backend/services/analytics/internal/repository"
)

type AnalyticsService struct {
	eventRepo   *repository.EventRepository
	videoStats  *repository.VideoStatsRepository
	channelStats *repository.ChannelStatsRepository
	liveStats   *repository.LiveStatsRepository
	counterRepo *repository.CounterRepository
}

func NewAnalyticsService(
	eventRepo *repository.EventRepository,
	videoStats *repository.VideoStatsRepository,
	channelStats *repository.ChannelStatsRepository,
	liveStats *repository.LiveStatsRepository,
	counterRepo *repository.CounterRepository,
) *AnalyticsService {
	return &AnalyticsService{
		eventRepo:    eventRepo,
		videoStats:   videoStats,
		channelStats: channelStats,
		liveStats:    liveStats,
		counterRepo:  counterRepo,
	}
}

// --- Request / Response DTOs ---

type IngestEventRequest struct {
	EventType  string `json:"event_type"`
	VideoID    uint64 `json:"video_id"`
	ChannelID  uint64 `json:"channel_id"`
	Properties string `json:"properties"`
	SessionID  string `json:"session_id"`
}

type DateRangeRequest struct {
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
}

type VideoAnalyticsResponse struct {
	VideoID    uint64                   `json:"video_id"`
	StartDate  string                   `json:"start_date"`
	EndDate    string                   `json:"end_date"`
	DailyStats []model.VideoDailyStats  `json:"daily_stats"`
}

type RealtimeVideoResponse struct {
	VideoID      uint64 `json:"video_id"`
	ViewsLast5m  int64  `json:"views_last_5m"`
	ViewsLast60m int64  `json:"views_last_60m"`
}

type ChannelAnalyticsResponse struct {
	ChannelID  uint64                     `json:"channel_id"`
	StartDate  string                     `json:"start_date"`
	EndDate    string                     `json:"end_date"`
	DailyStats []model.ChannelDailyStats  `json:"daily_stats"`
}

type ChannelOverviewResponse struct {
	ChannelID        uint64 `json:"channel_id"`
	TotalViews       int64  `json:"total_views"`
	TotalWatchTime   int64  `json:"total_watch_time_seconds"`
	NewSubscribers   int64  `json:"new_subscribers"`
	LostSubscribers  int64  `json:"lost_subscribers"`
	TotalRevenueCents int64 `json:"total_revenue_cents"`
}

type IncrementCounterRequest struct {
	EntityType  string `json:"entity_type"`
	EntityID    uint64 `json:"entity_id"`
	CounterName string `json:"counter_name"`
	Delta       int64  `json:"delta"`
}

// --- Event Ingestion ---

func (s *AnalyticsService) IngestEvent(req IngestEventRequest, userID uint64, ipAddress, userAgent string) (*model.AnalyticsEvent, error) {
	if req.EventType == "" {
		return nil, errors.ErrValidation.WithMessage("event_type is required")
	}

	if !model.ValidEventTypes[model.EventType(req.EventType)] {
		return nil, errors.ErrValidation.WithMessage("invalid event_type: must be one of view, watch_time, like, share, search, click")
	}

	// Determine device type from user agent (simple heuristic).
	deviceType := detectDeviceType(userAgent)

	event := &model.AnalyticsEvent{
		EventType:  req.EventType,
		VideoID:    req.VideoID,
		ChannelID:  req.ChannelID,
		UserID:     userID,
		SessionID:  req.SessionID,
		Properties: req.Properties,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		DeviceType: deviceType,
	}

	if err := s.eventRepo.Create(event); err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to store analytics event")
	}

	// If this is a view event and we have a video ID, increment today's video stats.
	if req.EventType == string(model.EventTypeView) && req.VideoID > 0 {
		today := time.Now().Format("2006-01-02")
		_ = s.videoStats.IncrementColumn(req.VideoID, today, "views", 1)
	}

	return event, nil
}

// --- Video Analytics ---

func (s *AnalyticsService) GetVideoAnalytics(videoID uint64, dateRange DateRangeRequest) (*VideoAnalyticsResponse, error) {
	if videoID == 0 {
		return nil, errors.ErrValidation.WithMessage("video_id is required")
	}

	stats, err := s.videoStats.FindByVideoID(videoID, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to retrieve video analytics")
	}

	return &VideoAnalyticsResponse{
		VideoID:    videoID,
		StartDate:  dateRange.StartDate,
		EndDate:    dateRange.EndDate,
		DailyStats: stats,
	}, nil
}

func (s *AnalyticsService) GetVideoRealtime(videoID uint64) (*RealtimeVideoResponse, error) {
	if videoID == 0 {
		return nil, errors.ErrValidation.WithMessage("video_id is required")
	}

	views5m, err := s.eventRepo.CountRecentByVideo(videoID, 5)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to retrieve realtime stats")
	}

	views60m, err := s.eventRepo.CountRecentByVideo(videoID, 60)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to retrieve realtime stats")
	}

	return &RealtimeVideoResponse{
		VideoID:      videoID,
		ViewsLast5m:  views5m,
		ViewsLast60m: views60m,
	}, nil
}

// --- Channel Analytics ---

func (s *AnalyticsService) GetChannelAnalytics(channelID uint64, dateRange DateRangeRequest) (*ChannelAnalyticsResponse, error) {
	if channelID == 0 {
		return nil, errors.ErrValidation.WithMessage("channel_id is required")
	}

	stats, err := s.channelStats.FindByChannelID(channelID, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to retrieve channel analytics")
	}

	return &ChannelAnalyticsResponse{
		ChannelID:  channelID,
		StartDate:  dateRange.StartDate,
		EndDate:    dateRange.EndDate,
		DailyStats: stats,
	}, nil
}

func (s *AnalyticsService) GetChannelOverview(channelID uint64) (*ChannelOverviewResponse, error) {
	if channelID == 0 {
		return nil, errors.ErrValidation.WithMessage("channel_id is required")
	}

	stats, err := s.channelStats.SumByChannelID(channelID)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to retrieve channel overview")
	}

	return &ChannelOverviewResponse{
		ChannelID:         channelID,
		TotalViews:        stats.Views,
		TotalWatchTime:    stats.WatchTimeSeconds,
		NewSubscribers:    stats.NewSubscribers,
		LostSubscribers:   stats.LostSubscribers,
		TotalRevenueCents: stats.RevenueCents,
	}, nil
}

// --- Live Session Stats ---

func (s *AnalyticsService) GetLiveSessionStats(sessionID uint64) (*model.LiveSessionStats, error) {
	if sessionID == 0 {
		return nil, errors.ErrValidation.WithMessage("session_id is required")
	}

	stats, err := s.liveStats.FindBySessionID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("live session stats not found")
		}
		return nil, errors.ErrInternal.WithMessage("failed to retrieve live session stats")
	}
	return stats, nil
}

// --- Counter Operations ---

func (s *AnalyticsService) IncrementCounter(req IncrementCounterRequest) (*model.CounterSnapshot, error) {
	if req.EntityType == "" {
		return nil, errors.ErrValidation.WithMessage("entity_type is required")
	}
	if req.EntityID == 0 {
		return nil, errors.ErrValidation.WithMessage("entity_id is required")
	}
	if req.CounterName == "" {
		return nil, errors.ErrValidation.WithMessage("counter_name is required")
	}
	if req.Delta == 0 {
		return nil, errors.ErrValidation.WithMessage("delta must be non-zero")
	}

	snapshot, err := s.counterRepo.Increment(req.EntityType, req.EntityID, req.CounterName, req.Delta)
	if err != nil {
		return nil, errors.ErrInternal.WithMessage("failed to increment counter")
	}
	return snapshot, nil
}

// --- Daily Aggregation ---

func (s *AnalyticsService) RunDailyAggregation() error {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Aggregate video stats from events for yesterday.
	if err := s.aggregateVideoStats(yesterday); err != nil {
		return errors.ErrInternal.WithMessage("failed to aggregate video stats: " + err.Error())
	}

	// Aggregate channel stats from events for yesterday.
	if err := s.aggregateChannelStats(yesterday); err != nil {
		return errors.ErrInternal.WithMessage("failed to aggregate channel stats: " + err.Error())
	}

	return nil
}

func (s *AnalyticsService) aggregateVideoStats(date string) error {
	// Get distinct video IDs that had events on the given date.
	events, err := s.eventRepo.CountByVideoAndDate(0, date, date)
	if err != nil {
		return err
	}

	// Collect unique video IDs from events.
	videoIDs := make(map[uint64]bool)
	for _, e := range events {
		if e.VideoID > 0 {
			videoIDs[e.VideoID] = true
		}
	}

	// For each video, count events by type and upsert stats.
	for videoID := range videoIDs {
		videoEvents, err := s.eventRepo.CountByVideoAndDate(videoID, date, date)
		if err != nil {
			continue
		}

		stats := &model.VideoDailyStats{
			VideoID: videoID,
			Date:    date,
		}

		uniqueUsers := make(map[uint64]bool)
		for _, e := range videoEvents {
			switch model.EventType(e.EventType) {
			case model.EventTypeView:
				stats.Views++
			case model.EventTypeLike:
				stats.Likes++
			case model.EventTypeShare:
				stats.Shares++
			}
			if e.UserID > 0 {
				uniqueUsers[e.UserID] = true
			}
		}
		stats.UniqueViewers = int64(len(uniqueUsers))

		_ = s.videoStats.Upsert(stats)
	}

	return nil
}

func (s *AnalyticsService) aggregateChannelStats(date string) error {
	events, err := s.eventRepo.CountByChannelAndDate(0, date, date)
	if err != nil {
		return err
	}

	channelIDs := make(map[uint64]bool)
	for _, e := range events {
		if e.ChannelID > 0 {
			channelIDs[e.ChannelID] = true
		}
	}

	for channelID := range channelIDs {
		channelEvents, err := s.eventRepo.CountByChannelAndDate(channelID, date, date)
		if err != nil {
			continue
		}

		stats := &model.ChannelDailyStats{
			ChannelID: channelID,
			Date:      date,
		}

		for _, e := range channelEvents {
			switch model.EventType(e.EventType) {
			case model.EventTypeView:
				stats.Views++
			}
		}

		_ = s.channelStats.Upsert(stats)
	}

	return nil
}

// --- Helpers ---

func detectDeviceType(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}
	// Simple heuristic based on common user agent strings.
	for _, keyword := range []string{"Mobile", "Android", "iPhone"} {
		if containsIgnoreCase(userAgent, keyword) {
			return "mobile"
		}
	}
	for _, keyword := range []string{"iPad", "Tablet"} {
		if containsIgnoreCase(userAgent, keyword) {
			return "tablet"
		}
	}
	return "desktop"
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	subLower := toLower(substr)
	return len(sLower) >= len(subLower) && containsStr(sLower, subLower)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
