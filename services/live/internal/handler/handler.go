package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/live/internal/service"
)

type Handler struct {
	liveService *service.LiveService
}

func New(liveService *service.LiveService) *Handler {
	return &Handler{liveService: liveService}
}

// ── Room CRUD ───────────────────────────────────────────────────────────────

func (h *Handler) CreateRoom(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req service.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	room, err := h.liveService.CreateRoom(userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	// Include stream_key in the creation response
	return response.Created(c, fiber.Map{
		"room":       room,
		"stream_key": room.StreamKey,
	})
}

func (h *Handler) GetRoom(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room id"))
	}
	room, sErr := h.liveService.GetRoom(id)
	if sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, room)
}

func (h *Handler) UpdateRoom(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room id"))
	}
	userID := middleware.GetUserID(c)
	var req service.UpdateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	room, sErr := h.liveService.UpdateRoom(id, userID, req)
	if sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, room)
}

func (h *Handler) DeleteRoom(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room id"))
	}
	userID := middleware.GetUserID(c)
	if sErr := h.liveService.DeleteRoom(id, userID); sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

// ── Stream key ──────────────────────────────────────────────────────────────

func (h *Handler) RegenerateStreamKey(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room id"))
	}
	userID := middleware.GetUserID(c)
	room, sErr := h.liveService.RegenerateStreamKey(id, userID)
	if sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{
		"stream_key": room.StreamKey,
	})
}

// ── Live stream lifecycle ───────────────────────────────────────────────────

func (h *Handler) GoLive(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room id"))
	}
	userID := middleware.GetUserID(c)
	room, sErr := h.liveService.GoLive(id, userID)
	if sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, room)
}

func (h *Handler) EndStream(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room id"))
	}
	userID := middleware.GetUserID(c)
	room, sErr := h.liveService.EndStream(id, userID)
	if sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, room)
}

// ── Playback ────────────────────────────────────────────────────────────────

func (h *Handler) GetPlaybackInfo(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room id"))
	}
	room, sErr := h.liveService.GetPlaybackInfo(id)
	if sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{
		"playback_url":  room.PlaybackURL,
		"status":        room.Status,
		"viewer_count":  room.ViewerCount,
		"thumbnail_url": room.ThumbnailURL,
	})
}

// ── Listings ────────────────────────────────────────────────────────────────

func (h *Handler) ListLiveRooms(c *fiber.Ctx) error {
	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}
	pg.Normalize()

	rooms, total, err := h.liveService.ListLiveRooms(pg.Offset(), pg.PageSize)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Paginated(c, rooms, types.NewPaginationMeta(pg, total))
}

func (h *Handler) ListChannelRooms(c *fiber.Ctx) error {
	channelID, err := strconv.ParseUint(c.Params("channelId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel id"))
	}

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}
	pg.Normalize()

	rooms, total, sErr := h.liveService.ListChannelRooms(channelID, pg.Offset(), pg.PageSize)
	if sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Paginated(c, rooms, types.NewPaginationMeta(pg, total))
}

// ── Stream auth callback ────────────────────────────────────────────────────

func (h *Handler) StreamAuth(c *fiber.Ctx) error {
	var req service.StreamAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	room, err := h.liveService.StreamAuth(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{
		"room_id":    room.ID,
		"channel_id": room.ChannelID,
		"status":     room.Status,
	})
}

// ── Sessions ────────────────────────────────────────────────────────────────

func (h *Handler) ListSessions(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid room id"))
	}

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}
	pg.Normalize()

	sessions, total, sErr := h.liveService.ListSessions(roomID, pg.Offset(), pg.PageSize)
	if sErr != nil {
		if appErr, ok := sErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Paginated(c, sessions, types.NewPaginationMeta(pg, total))
}
