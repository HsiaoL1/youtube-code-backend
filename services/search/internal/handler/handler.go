package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/search/internal/service"
)

// Handler contains search endpoint handlers.
type Handler struct {
	svc *service.SearchService
}

// New creates a new handler instance.
func New(svc *service.SearchService) *Handler {
	return &Handler{svc: svc}
}

// SearchVideos handles GET /videos?q=&category=&sort= — search videos.
func (h *Handler) SearchVideos(c *fiber.Ctx) error {
	query := c.Query("q")
	category := c.Query("category")
	sort := c.Query("sort")

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	videos, meta, err := h.svc.SearchVideos(query, category, sort, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, videos, meta)
}

// SearchChannels handles GET /channels?q= — search channels.
func (h *Handler) SearchChannels(c *fiber.Ctx) error {
	query := c.Query("q")

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	channels, meta, err := h.svc.SearchChannels(query, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, channels, meta)
}

// SearchLive handles GET /live?q= — search live rooms.
func (h *Handler) SearchLive(c *fiber.Ctx) error {
	query := c.Query("q")

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	rooms, meta, err := h.svc.SearchLive(query, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, rooms, meta)
}

// SearchAll handles GET /all?q= — combined search across videos, channels, live.
func (h *Handler) SearchAll(c *fiber.Ctx) error {
	query := c.Query("q")

	result, err := h.svc.SearchAll(query)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, result)
}

// IndexVideo handles POST /index/videos — upsert video index entry.
func (h *Handler) IndexVideo(c *fiber.Ctx) error {
	var req service.UpsertVideoRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	video, err := h.svc.UpsertVideo(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, video)
}

// DeleteVideoIndex handles DELETE /index/videos/:videoId — remove video from index.
func (h *Handler) DeleteVideoIndex(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video ID"))
	}

	if err := h.svc.DeleteVideo(videoID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.NoContent(c)
}

// IndexChannel handles POST /index/channels — upsert channel index entry.
func (h *Handler) IndexChannel(c *fiber.Ctx) error {
	var req service.UpsertChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	channel, err := h.svc.UpsertChannel(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, channel)
}

// IndexLive handles POST /index/live — upsert live room index entry.
func (h *Handler) IndexLive(c *fiber.Ctx) error {
	var req service.UpsertLiveRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	live, err := h.svc.UpsertLive(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, live)
}
