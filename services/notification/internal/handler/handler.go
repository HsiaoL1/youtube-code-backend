package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/notification/internal/service"
)

type Handler struct {
	notifService *service.NotificationService
}

func New(ns *service.NotificationService) *Handler {
	return &Handler{notifService: ns}
}

// ListNotifications returns paginated notifications for the authenticated user.
func (h *Handler) ListNotifications(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination params"))
	}
	pg.Normalize()

	notifications, total, err := h.notifService.List(userID, pg.Offset(), pg.PageSize)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Paginated(c, notifications, types.NewPaginationMeta(pg, total))
}

// UnreadCount returns the number of unread notifications.
func (h *Handler) UnreadCount(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	count, err := h.notifService.UnreadCount(userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"unread_count": count})
}

// MarkAsRead marks a single notification as read.
func (h *Handler) MarkAsRead(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid notification id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.notifService.MarkAsRead(id, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "marked as read"})
}

// MarkAllAsRead marks all notifications as read for the authenticated user.
func (h *Handler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	if svcErr := h.notifService.MarkAllAsRead(userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "all notifications marked as read"})
}

// DeleteNotification deletes a notification.
func (h *Handler) DeleteNotification(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid notification id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.notifService.Delete(id, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

// GetPreferences returns notification preferences for the authenticated user.
func (h *Handler) GetPreferences(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	pref, err := h.notifService.GetPreferences(userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, pref)
}

// UpdatePreferences updates notification preferences for the authenticated user.
func (h *Handler) UpdatePreferences(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req service.UpdatePreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	pref, err := h.notifService.UpdatePreferences(userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, pref)
}

// SendNotification is an internal endpoint to send a notification to a user.
func (h *Handler) SendNotification(c *fiber.Ctx) error {
	var req service.SendNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	n, err := h.notifService.Send(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, n)
}

// BroadcastNotification is an internal endpoint to send a notification to multiple users.
func (h *Handler) BroadcastNotification(c *fiber.Ctx) error {
	var req service.BroadcastNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	count, err := h.notifService.Broadcast(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, fiber.Map{"sent": count})
}
