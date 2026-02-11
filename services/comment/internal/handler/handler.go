package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/comment/internal/service"
)

type Handler struct {
	commentService       *service.CommentService
	sensitiveWordService *service.SensitiveWordService
}

func New(cs *service.CommentService, sws *service.SensitiveWordService) *Handler {
	return &Handler{commentService: cs, sensitiveWordService: sws}
}

// ---------- Comment endpoints ----------

func (h *Handler) ListByVideo(c *fiber.Ctx) error {
	videoID, err := strconv.ParseUint(c.Params("videoId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid video id"))
	}

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}
	pg.Normalize()

	userID := middleware.GetUserID(c)

	comments, total, svcErr := h.commentService.ListByVideoID(videoID, pg.Offset(), pg.PageSize, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, comments, types.NewPaginationMeta(pg, total))
}

func (h *Handler) ListReplies(c *fiber.Ctx) error {
	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid comment id"))
	}

	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination parameters"))
	}
	pg.Normalize()

	userID := middleware.GetUserID(c)

	replies, total, svcErr := h.commentService.ListReplies(commentID, pg.Offset(), pg.PageSize, userID)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}

	return response.Paginated(c, replies, types.NewPaginationMeta(pg, total))
}

func (h *Handler) CreateComment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req service.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	comment, svcErr := h.commentService.Create(userID, req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, comment)
}

func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid comment id"))
	}

	if svcErr := h.commentService.Delete(userID, commentID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) LikeComment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid comment id"))
	}

	if svcErr := h.commentService.Like(userID, commentID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) UnlikeComment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid comment id"))
	}

	if svcErr := h.commentService.Unlike(userID, commentID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) PinComment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid comment id"))
	}

	if svcErr := h.commentService.Pin(userID, commentID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) HeartComment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid comment id"))
	}

	if svcErr := h.commentService.Heart(userID, commentID); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) ReportComment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid comment id"))
	}

	var req service.ReportCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	if svcErr := h.commentService.Report(userID, commentID, req.Reason); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

// ---------- Sensitive word endpoints ----------

func (h *Handler) ListSensitiveWords(c *fiber.Ctx) error {
	words, svcErr := h.sensitiveWordService.List()
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, words)
}

func (h *Handler) AddSensitiveWord(c *fiber.Ctx) error {
	var req service.AddSensitiveWordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	word, svcErr := h.sensitiveWordService.Add(req)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, word)
}

func (h *Handler) DeleteSensitiveWord(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid sensitive word id"))
	}

	if svcErr := h.sensitiveWordService.Delete(id); svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}
