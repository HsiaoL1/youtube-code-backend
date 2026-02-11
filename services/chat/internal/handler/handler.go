package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/chat/internal/service"
)

// Handler contains chat endpoint handlers.
type Handler struct {
	svc *service.ChatService
}

// New creates a new handler instance.
func New(svc *service.ChatService) *Handler {
	return &Handler{svc: svc}
}

// GetMessages handles GET /messages/:roomId — get paginated message history.
func (h *Handler) GetMessages(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}

	messages, meta, err := h.svc.GetMessages(roomID, pg)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, messages, meta)
}

// SendMessage handles POST /rooms/:roomId/send — send a message to a room.
func (h *Handler) SendMessage(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	var req service.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	msg, err := h.svc.SendMessage(roomID, userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Created(c, msg)
}

// GetRoomConfig handles GET /rooms/:roomId/config — get room configuration.
func (h *Handler) GetRoomConfig(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	cfg, err := h.svc.GetRoomConfig(roomID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, cfg)
}

// UpdateRoomConfig handles PUT /rooms/:roomId/config — update room configuration.
func (h *Handler) UpdateRoomConfig(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	var req service.UpdateRoomConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	cfg, err := h.svc.UpdateRoomConfig(roomID, userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, cfg)
}

// ListManagers handles GET /rooms/:roomId/managers — list room managers.
func (h *Handler) ListManagers(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	managers, err := h.svc.ListManagers(roomID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, managers)
}

// AddManager handles POST /rooms/:roomId/managers — add a manager to a room.
func (h *Handler) AddManager(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	var req service.AddManagerRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	mgr, err := h.svc.AddManager(roomID, userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Created(c, mgr)
}

// RemoveManager handles DELETE /rooms/:roomId/managers/:userId — remove a manager.
func (h *Handler) RemoveManager(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	targetUserID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid user ID"))
	}

	if err := h.svc.RemoveManager(roomID, userID, targetUserID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.NoContent(c)
}

// MuteUser handles POST /rooms/:roomId/mute — mute a user in a room.
func (h *Handler) MuteUser(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	var req service.MuteUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	mute, err := h.svc.MuteUser(roomID, userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Created(c, mute)
}

// UnmuteUser handles DELETE /rooms/:roomId/mute/:userId — unmute a user.
func (h *Handler) UnmuteUser(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	targetUserID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid user ID"))
	}

	if err := h.svc.UnmuteUser(roomID, userID, targetUserID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.NoContent(c)
}

// ListFilters handles GET /rooms/:roomId/filters — list keyword filters.
func (h *Handler) ListFilters(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	filters, err := h.svc.ListFilters(roomID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.OK(c, filters)
}

// AddFilter handles POST /rooms/:roomId/filters — add a keyword filter.
func (h *Handler) AddFilter(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	var req service.AddKeywordFilterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	filter, err := h.svc.AddFilter(roomID, userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Created(c, filter)
}

// RemoveFilter handles DELETE /rooms/:roomId/filters/:id — remove a keyword filter.
func (h *Handler) RemoveFilter(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Err(c, errors.ErrUnauthorized)
	}

	roomID, err := strconv.ParseUint(c.Params("roomId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room ID"))
	}

	filterID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid filter ID"))
	}

	if err := h.svc.RemoveFilter(roomID, userID, filterID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.NoContent(c)
}
