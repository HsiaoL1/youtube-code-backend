package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/social-graph/internal/service"
)

// Handler contains social-graph endpoint handlers.
type Handler struct {
	svc *service.SubscriptionService
}

// New creates a new handler instance.
func New(svc *service.SubscriptionService) *Handler {
	return &Handler{svc: svc}
}

// Subscribe handles POST /subscribe — subscribe to a channel.
func (h *Handler) Subscribe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	var req service.SubscribeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	sub, err := h.svc.Subscribe(userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Created(c, sub)
}

// Unsubscribe handles DELETE /subscribe/:channelId — unsubscribe from a channel.
func (h *Handler) Unsubscribe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel ID"))
	}

	if err := h.svc.Unsubscribe(userID, channelID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.NoContent(c)
}

// ListSubscriptions handles GET /subscriptions — list my subscriptions.
func (h *Handler) ListSubscriptions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	subs, meta, err := h.svc.ListSubscriptions(userID, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, subs, meta)
}

// ListSubscribers handles GET /subscribers/:channelId — list channel subscribers.
func (h *Handler) ListSubscribers(c *fiber.Ctx) error {
	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel ID"))
	}

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	subs, meta, err := h.svc.ListSubscribers(channelID, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, subs, meta)
}

// GetSubscriberCount handles GET /subscribers/:channelId/count — get subscriber count.
func (h *Handler) GetSubscriberCount(c *fiber.Ctx) error {
	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel ID"))
	}

	count, err := h.svc.GetSubscriberCount(channelID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, fiber.Map{"channel_id": channelID, "subscriber_count": count})
}

// CheckSubscription handles GET /subscriptions/check/:channelId — check if subscribed.
func (h *Handler) CheckSubscription(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel ID"))
	}

	sub, subscribed, err := h.svc.CheckSubscription(userID, channelID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	result := fiber.Map{
		"subscribed": subscribed,
		"channel_id": channelID,
	}
	if subscribed && sub != nil {
		result["subscription"] = sub
	}

	return response.OK(c, result)
}

// UpdateNotifyPreference handles PUT /subscribe/:channelId/notify — update notification preference.
func (h *Handler) UpdateNotifyPreference(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel ID"))
	}

	var req service.UpdateNotifyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	sub, err := h.svc.UpdateNotifyPreference(userID, channelID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, sub)
}

// GetFollowerIDs handles GET /followers/:channelId/ids — internal endpoint for follower IDs.
func (h *Handler) GetFollowerIDs(c *fiber.Ctx) error {
	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel ID"))
	}

	ids, err := h.svc.GetFollowerIDs(channelID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, fiber.Map{"channel_id": channelID, "follower_ids": ids})
}
