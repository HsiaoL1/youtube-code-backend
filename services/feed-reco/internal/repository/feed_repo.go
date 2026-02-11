package repository

import (
	"youtube-code-backend/services/feed-reco/internal/model"

	"gorm.io/gorm"
)

// FeedRepository handles all feed-related database queries.
type FeedRepository struct {
	db *gorm.DB
}

// NewFeedRepository creates a new FeedRepository.
func NewFeedRepository(db *gorm.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

// HomeFeed returns popular public videos ordered by trending score (24h).
// If userID > 0, results can be personalized in the future.
func (r *FeedRepository) HomeFeed(offset, limit int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	baseQuery := r.db.Table("videos").
		Joins("LEFT JOIN trending_scores ON trending_scores.video_id = videos.id").
		Where("videos.status = ? AND videos.visibility = ?", model.VideoStatusReady, model.VideoVisibilityPublic)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Select("videos.*, trending_scores.score_24h, trending_scores.category").
		Order("COALESCE(trending_scores.score_24h, 0) DESC, videos.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&videos).Error

	return videos, total, err
}

// SubscriptionFeed returns recent public videos from channels the user subscribes to.
func (r *FeedRepository) SubscriptionFeed(userID uint64, offset, limit int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	baseQuery := r.db.Table("videos").
		Joins("INNER JOIN subscriptions ON subscriptions.channel_id = videos.channel_id").
		Where("subscriptions.subscriber_id = ? AND videos.status = ? AND videos.visibility = ?",
			userID, model.VideoStatusReady, model.VideoVisibilityPublic)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Select("videos.*").
		Order("videos.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&videos).Error

	return videos, total, err
}

// TrendingVideos returns trending videos, optionally filtered by category.
func (r *FeedRepository) TrendingVideos(category string, offset, limit int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	baseQuery := r.db.Table("videos").
		Joins("INNER JOIN trending_scores ON trending_scores.video_id = videos.id").
		Where("videos.status = ? AND videos.visibility = ?", model.VideoStatusReady, model.VideoVisibilityPublic)

	if category != "" {
		baseQuery = baseQuery.Where("trending_scores.category = ?", category)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Select("videos.*, trending_scores.score_24h, trending_scores.category").
		Order("trending_scores.score_7d DESC").
		Offset(offset).Limit(limit).
		Find(&videos).Error

	return videos, total, err
}

// CategoryFeed returns videos for a specific category using the precomputed category_feeds table.
func (r *FeedRepository) CategoryFeed(category string, offset, limit int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	baseQuery := r.db.Table("videos").
		Joins("INNER JOIN category_feeds ON category_feeds.video_id = videos.id").
		Where("category_feeds.category = ? AND videos.status = ? AND videos.visibility = ?",
			category, model.VideoStatusReady, model.VideoVisibilityPublic)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Select("videos.*, category_feeds.score, category_feeds.category").
		Order("category_feeds.position ASC").
		Offset(offset).Limit(limit).
		Find(&videos).Error

	return videos, total, err
}

// ShortsFeed returns short-form videos ordered by trending score.
func (r *FeedRepository) ShortsFeed(offset, limit int) ([]model.Video, int64, error) {
	var videos []model.Video
	var total int64

	baseQuery := r.db.Table("videos").
		Joins("LEFT JOIN trending_scores ON trending_scores.video_id = videos.id").
		Where("videos.type = ? AND videos.status = ? AND videos.visibility = ?",
			model.VideoTypeShort, model.VideoStatusReady, model.VideoVisibilityPublic)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := baseQuery.
		Select("videos.*, trending_scores.score_24h").
		Order("COALESCE(trending_scores.score_24h, 0) DESC, videos.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&videos).Error

	return videos, total, err
}

// RelatedVideos returns videos related to the given videoID by same channel or category.
func (r *FeedRepository) RelatedVideos(videoID uint64, limit int) ([]model.Video, error) {
	// First, fetch the source video to know its channel and category.
	var source model.Video
	if err := r.db.First(&source, videoID).Error; err != nil {
		return nil, err
	}

	// Look up category from trending_scores.
	var ts model.TrendingScore
	_ = r.db.Where("video_id = ?", videoID).First(&ts).Error
	category := ts.Category

	var videos []model.Video

	query := r.db.Table("videos").
		Joins("LEFT JOIN trending_scores ON trending_scores.video_id = videos.id").
		Where("videos.id != ? AND videos.status = ? AND videos.visibility = ?",
			videoID, model.VideoStatusReady, model.VideoVisibilityPublic)

	// Prefer same channel or same category.
	if category != "" {
		query = query.Where("videos.channel_id = ? OR trending_scores.category = ?", source.ChannelID, category)
	} else {
		query = query.Where("videos.channel_id = ?", source.ChannelID)
	}

	err := query.
		Select("videos.*").
		Order("COALESCE(trending_scores.score_24h, 0) DESC, videos.view_count DESC").
		Limit(limit).
		Find(&videos).Error

	return videos, err
}
