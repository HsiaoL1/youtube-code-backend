package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/video/internal/service"
)

type Handler struct {
	videoService    *service.VideoService
	playlistService *service.PlaylistService
}

func New(vs *service.VideoService, ps *service.PlaylistService) *Handler {
	return &Handler{videoService: vs, playlistService: ps}
}

// --- Video endpoints ---

func (h *Handler) CreateVideo(c *fiber.Ctx) error {
	var req service.CreateVideoRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	video, err := h.videoService.Create(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, video)
}

func (h *Handler) GetVideo(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	video, svcErr := h.videoService.GetByID(id)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, video)
}

func (h *Handler) UpdateVideo(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	var req service.UpdateVideoRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	// The channel_id is passed in the body to verify ownership.
	var body struct {
		ChannelID uint64 `json:"channel_id"`
	}
	_ = c.BodyParser(&body)

	video, svcErr := h.videoService.Update(id, body.ChannelID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, video)
}

func (h *Handler) DeleteVideo(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	var body struct {
		ChannelID uint64 `json:"channel_id"`
	}
	_ = c.BodyParser(&body)

	if svcErr := h.videoService.Delete(id, body.ChannelID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) PublishVideo(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	var body struct {
		ChannelID uint64 `json:"channel_id"`
	}
	_ = c.BodyParser(&body)

	video, svcErr := h.videoService.Publish(id, body.ChannelID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, video)
}

// --- Like / Dislike ---

func (h *Handler) LikeVideo(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.videoService.LikeVideo(id, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "liked"})
}

func (h *Handler) DislikeVideo(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.videoService.DislikeVideo(id, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "disliked"})
}

func (h *Handler) RemoveLike(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.videoService.RemoveLike(id, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

// --- Favorites ---

func (h *Handler) AddFavorite(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.videoService.AddFavorite(id, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, fiber.Map{"message": "added to favorites"})
}

func (h *Handler) RemoveFavorite(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.videoService.RemoveFavorite(id, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

// --- Watch Progress / History ---

func (h *Handler) GetProgress(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	userID := middleware.GetUserID(c)
	wh, svcErr := h.videoService.GetProgress(id, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, wh)
}

func (h *Handler) UpdateProgress(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	var req service.UpdateProgressRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	userID := middleware.GetUserID(c)
	wh, svcErr := h.videoService.UpdateProgress(id, userID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, wh)
}

func (h *Handler) GetWatchHistory(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination params"))
	}
	pg.Normalize()

	items, total, err := h.videoService.GetWatchHistory(userID, pg.Offset(), pg.PageSize)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Paginated(c, items, types.NewPaginationMeta(pg, total))
}

// --- Playlist endpoints ---

func (h *Handler) CreatePlaylist(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req service.CreatePlaylistRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	playlist, err := h.playlistService.Create(userID, req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, playlist)
}

func (h *Handler) GetPlaylist(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid playlist id"))
	}

	playlist, svcErr := h.playlistService.GetByID(id)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, playlist)
}

func (h *Handler) UpdatePlaylist(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid playlist id"))
	}

	userID := middleware.GetUserID(c)

	var req service.UpdatePlaylistRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	playlist, svcErr := h.playlistService.Update(id, userID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, playlist)
}

func (h *Handler) DeletePlaylist(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid playlist id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.playlistService.Delete(id, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) AddPlaylistItem(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid playlist id"))
	}

	userID := middleware.GetUserID(c)

	var req service.AddPlaylistItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	item, svcErr := h.playlistService.AddItem(id, userID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, item)
}

func (h *Handler) RemovePlaylistItem(c *fiber.Ctx) error {
	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid playlist id"))
	}

	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	userID := middleware.GetUserID(c)
	if svcErr := h.playlistService.RemoveItem(playlistID, videoID, userID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}
