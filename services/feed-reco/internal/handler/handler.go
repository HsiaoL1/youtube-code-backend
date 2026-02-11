package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/feed-reco/internal/service"
)

// Handler contains endpoint handlers for the feed-reco service.
type Handler struct {
	feedService *service.FeedService
}

// New creates a new Handler.
func New(feedService *service.FeedService) *Handler {
	return &Handler{feedService: feedService}
}

// HomeFeed returns the home feed, personalized if the user is authenticated.
func (h *Handler) HomeFeed(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	videos, meta, err := h.feedService.HomeFeed(userID, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, videos, meta)
}

// SubscriptionFeed returns videos from channels the authenticated user subscribes to.
func (h *Handler) SubscriptionFeed(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	videos, meta, err := h.feedService.SubscriptionFeed(userID, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, videos, meta)
}

// Trending returns trending videos with optional category filter.
func (h *Handler) Trending(c *fiber.Ctx) error {
	category := c.Query("category")

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	videos, meta, err := h.feedService.Trending(category, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, videos, meta)
}

// CategoryFeed returns videos for a specific category.
func (h *Handler) CategoryFeed(c *fiber.Ctx) error {
	category := c.Params("category")
	if category == "" {
		return response.Err(c, errors.ErrBadRequest.WithMessage("category is required"))
	}

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	videos, meta, err := h.feedService.CategoryFeed(category, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, videos, meta)
}

// ShortsFeed returns a feed of short-form videos.
func (h *Handler) ShortsFeed(c *fiber.Ctx) error {
	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	videos, meta, err := h.feedService.ShortsFeed(pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, videos, meta)
}

// RelatedVideos returns videos related to the given video ID.
func (h *Handler) RelatedVideos(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video ID"))
	}

	videos, err := h.feedService.RelatedVideos(videoID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, videos)
}
