package service

import (
	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/feed-reco/internal/model"
	"youtube-code-backend/services/feed-reco/internal/repository"

	"gorm.io/gorm"
)

// FeedService contains the business logic for feed and recommendation endpoints.
type FeedService struct {
	repo *repository.FeedRepository
}

// NewFeedService creates a new FeedService.
func NewFeedService(repo *repository.FeedRepository) *FeedService {
	return &FeedService{repo: repo}
}

// HomeFeed returns the home feed, optionally personalized for the given user.
func (s *FeedService) HomeFeed(userID uint64, pg types.PaginationRequest) ([]model.Video, types.PaginationMeta, error) {
	pg.Normalize()

	videos, total, err := s.repo.HomeFeed(pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to fetch home feed")
	}

	meta := types.NewPaginationMeta(pg, total)
	return videos, meta, nil
}

// SubscriptionFeed returns videos from channels the user subscribes to.
func (s *FeedService) SubscriptionFeed(userID uint64, pg types.PaginationRequest) ([]model.Video, types.PaginationMeta, error) {
	if userID == 0 {
		return nil, types.PaginationMeta{}, errors.ErrUnauthorized.WithMessage("authentication required for subscription feed")
	}

	pg.Normalize()

	videos, total, err := s.repo.SubscriptionFeed(userID, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to fetch subscription feed")
	}

	meta := types.NewPaginationMeta(pg, total)
	return videos, meta, nil
}

// Trending returns trending videos, optionally filtered by category.
func (s *FeedService) Trending(category string, pg types.PaginationRequest) ([]model.Video, types.PaginationMeta, error) {
	pg.Normalize()

	videos, total, err := s.repo.TrendingVideos(category, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to fetch trending videos")
	}

	meta := types.NewPaginationMeta(pg, total)
	return videos, meta, nil
}

// CategoryFeed returns videos for a specific category.
func (s *FeedService) CategoryFeed(category string, pg types.PaginationRequest) ([]model.Video, types.PaginationMeta, error) {
	if category == "" {
		return nil, types.PaginationMeta{}, errors.ErrBadRequest.WithMessage("category is required")
	}

	pg.Normalize()

	videos, total, err := s.repo.CategoryFeed(category, pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to fetch category feed")
	}

	meta := types.NewPaginationMeta(pg, total)
	return videos, meta, nil
}

// ShortsFeed returns short-form video feed.
func (s *FeedService) ShortsFeed(pg types.PaginationRequest) ([]model.Video, types.PaginationMeta, error) {
	pg.Normalize()

	videos, total, err := s.repo.ShortsFeed(pg.Offset(), pg.PageSize)
	if err != nil {
		return nil, types.PaginationMeta{}, errors.ErrInternal.WithMessage("failed to fetch shorts feed")
	}

	meta := types.NewPaginationMeta(pg, total)
	return videos, meta, nil
}

// RelatedVideos returns videos related to the given video.
func (s *FeedService) RelatedVideos(videoID uint64) ([]model.Video, error) {
	if videoID == 0 {
		return nil, errors.ErrBadRequest.WithMessage("video ID is required")
	}

	const relatedLimit = 20

	videos, err := s.repo.RelatedVideos(videoID, relatedLimit)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound.WithMessage("video not found")
		}
		return nil, errors.ErrInternal.WithMessage("failed to fetch related videos")
	}

	return videos, nil
}
