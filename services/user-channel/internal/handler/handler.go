package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/services/user-channel/internal/service"
)

type Handler struct {
	profileService *service.ProfileService
	channelService *service.ChannelService
}

func New(ps *service.ProfileService, cs *service.ChannelService) *Handler {
	return &Handler{profileService: ps, channelService: cs}
}

// ---------- Profile endpoints ----------

func (h *Handler) GetProfile(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid user id"))
	}

	profile, svcErr := h.profileService.GetByUserID(userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, profile)
}

func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req service.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	profile, svcErr := h.profileService.UpdateProfile(userID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, profile)
}

// ---------- Channel endpoints ----------

func (h *Handler) CreateChannel(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req service.CreateChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	channel, svcErr := h.channelService.Create(userID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, channel)
}

func (h *Handler) GetChannel(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel id"))
	}

	channel, svcErr := h.channelService.GetByID(id)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, channel)
}

func (h *Handler) GetChannelByHandle(c *fiber.Ctx) error {
	handle := c.Params("handle")
	if handle == "" {
		return response.Err(c, errors.ErrBadRequest.WithMessage("handle is required"))
	}

	channel, svcErr := h.channelService.GetByHandle(handle)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, channel)
}

func (h *Handler) UpdateChannel(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel id"))
	}

	var req service.UpdateChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	channel, svcErr := h.channelService.Update(userID, id, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, channel)
}

func (h *Handler) GetChannelStats(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid channel id"))
	}

	stats, svcErr := h.channelService.GetStats(id)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, stats)
}
