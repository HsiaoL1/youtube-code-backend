package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/services/analytics/internal/service"
)

type Handler struct {
	analyticsService *service.AnalyticsService
}

func New(as *service.AnalyticsService) *Handler {
	return &Handler{analyticsService: as}
}

// Ping is a simple liveness endpoint under the service domain.
func (h *Handler) Ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service": "analytics-service",
		"message": "pong",
	})
}

// IngestEvent handles POST /events — ingest analytics event (optional auth).
func (h *Handler) IngestEvent(c *fiber.Ctx) error {
	var req service.IngestEventRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	userID := middleware.GetUserID(c)
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	event, err := h.analyticsService.IngestEvent(req, userID, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, event)
}

// GetVideoAnalytics handles GET /videos/:videoId — get video analytics (auth).
func (h *Handler) GetVideoAnalytics(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	var dateRange service.DateRangeRequest
	if err := c.QueryParser(&dateRange); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid query params"))
	}

	result, svcErr := h.analyticsService.GetVideoAnalytics(videoID, dateRange)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, result)
}

// GetVideoRealtime handles GET /videos/:videoId/realtime — get real-time video stats.
func (h *Handler) GetVideoRealtime(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	result, svcErr := h.analyticsService.GetVideoRealtime(videoID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, result)
}

// GetChannelAnalytics handles GET /channels/:channelId — get channel analytics (auth).
func (h *Handler) GetChannelAnalytics(c *fiber.Ctx) error {
	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel id"))
	}

	var dateRange service.DateRangeRequest
	if err := c.QueryParser(&dateRange); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid query params"))
	}

	result, svcErr := h.analyticsService.GetChannelAnalytics(channelID, dateRange)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, result)
}

// GetChannelOverview handles GET /channels/:channelId/overview — channel overview (total stats).
func (h *Handler) GetChannelOverview(c *fiber.Ctx) error {
	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel id"))
	}

	result, svcErr := h.analyticsService.GetChannelOverview(channelID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, result)
}

// GetLiveSessionStats handles GET /live/:sessionId — get live session stats.
func (h *Handler) GetLiveSessionStats(c *fiber.Ctx) error {
	sessionID, err := strconv.ParseUint(c.Params("sessionId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid session id"))
	}

	result, svcErr := h.analyticsService.GetLiveSessionStats(sessionID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, result)
}

// IncrementCounter handles POST /counters/increment — internal: increment counter.
func (h *Handler) IncrementCounter(c *fiber.Ctx) error {
	var req service.IncrementCounterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	snapshot, err := h.analyticsService.IncrementCounter(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, snapshot)
}

// TriggerDailyAggregation handles POST /aggregate/daily — internal: trigger daily aggregation job (admin).
func (h *Handler) TriggerDailyAggregation(c *fiber.Ctx) error {
	if err := h.analyticsService.RunDailyAggregation(); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "daily aggregation completed"})
}
