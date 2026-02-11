package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/services/media/internal/service"
)

type Handler struct {
	uploadService *service.UploadService
	assetService  *service.AssetService
}

func New(uploadService *service.UploadService, assetService *service.AssetService) *Handler {
	return &Handler{
		uploadService: uploadService,
		assetService:  assetService,
	}
}

// --- Upload handlers ---

func (h *Handler) InitUpload(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req service.InitUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	session, err := h.uploadService.InitUpload(userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, session)
}

func (h *Handler) UploadPart(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return response.Err(c, errors.ErrBadRequest.WithMessage("session id is required"))
	}
	var req service.UploadPartRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	part, err := h.uploadService.UploadPart(userID, sessionID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, part)
}

func (h *Handler) CompleteUpload(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return response.Err(c, errors.ErrBadRequest.WithMessage("session id is required"))
	}
	session, err := h.uploadService.CompleteUpload(userID, sessionID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, session)
}

func (h *Handler) AbortUpload(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return response.Err(c, errors.ErrBadRequest.WithMessage("session id is required"))
	}
	if err := h.uploadService.AbortUpload(userID, sessionID); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) GetSessionStatus(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return response.Err(c, errors.ErrBadRequest.WithMessage("session id is required"))
	}
	session, parts, err := h.uploadService.GetSessionStatus(userID, sessionID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"session": session, "parts": parts})
}

// --- Asset handlers ---

func (h *Handler) GetAssetsByVideoID(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}
	assets, err := h.assetService.GetAssetsByVideoID(videoID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, assets)
}

func (h *Handler) GetVariantsByAssetID(c *fiber.Ctx) error {
	assetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid asset id"))
	}
	variants, err := h.assetService.GetVariantsByAssetID(assetID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, variants)
}

// --- Thumbnail handlers ---

func (h *Handler) GetThumbnailsByVideoID(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}
	thumbnails, err := h.assetService.GetThumbnailsByVideoID(videoID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, thumbnails)
}

func (h *Handler) AddThumbnail(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}
	var req service.AddThumbnailRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	thumbnail, err := h.assetService.AddThumbnail(videoID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, thumbnail)
}

// --- Subtitle handlers ---

func (h *Handler) GetSubtitlesByVideoID(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}
	subtitles, err := h.assetService.GetSubtitlesByVideoID(videoID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, subtitles)
}

func (h *Handler) AddSubtitle(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}
	var req service.AddSubtitleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	subtitle, err := h.assetService.AddSubtitle(videoID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, subtitle)
}

func (h *Handler) DeleteSubtitle(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid subtitle id"))
	}
	if err := h.assetService.DeleteSubtitle(id); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}
