package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/pkg/common/types"
	"youtube-code-backend/services/moderation/internal/service"
)

type Handler struct {
	svc *service.ModerationService
}

func New(svc *service.ModerationService) *Handler {
	return &Handler{svc: svc}
}

// --- Moderation Queue ---

func (h *Handler) ListQueue(c *fiber.Ctx) error {
	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination params"))
	}
	pg.Normalize()

	contentType := c.Query("content_type")

	items, total, err := h.svc.ListQueue(contentType, pg.Offset(), pg.PageSize)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Paginated(c, items, types.NewPaginationMeta(pg, total))
}

func (h *Handler) GetItem(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid item id"))
	}

	item, svcErr := h.svc.GetItem(id)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, item)
}

func (h *Handler) ApproveItem(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid item id"))
	}

	userID := middleware.GetUserID(c)
	item, svcErr := h.svc.ApproveItem(id, userID, c.IP())
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, item)
}

func (h *Handler) RejectItem(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid item id"))
	}

	var req service.RejectItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	userID := middleware.GetUserID(c)
	item, svcErr := h.svc.RejectItem(id, userID, req, c.IP())
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, item)
}

// --- Reports ---

func (h *Handler) CreateReport(c *fiber.Ctx) error {
	var req service.CreateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	userID := middleware.GetUserID(c)
	report, svcErr := h.svc.CreateReport(userID, req, c.IP())
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, report)
}

func (h *Handler) ListReports(c *fiber.Ctx) error {
	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination params"))
	}
	pg.Normalize()

	status := c.Query("status")

	reports, total, err := h.svc.ListReports(status, pg.Offset(), pg.PageSize)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Paginated(c, reports, types.NewPaginationMeta(pg, total))
}

func (h *Handler) UpdateReportStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid report id"))
	}

	var req service.UpdateReportStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	userID := middleware.GetUserID(c)
	report, svcErr := h.svc.UpdateReportStatus(id, userID, req, c.IP())
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, report)
}

// --- Enforcement Actions ---

func (h *Handler) CreateEnforcement(c *fiber.Ctx) error {
	var req service.CreateEnforcementRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}

	userID := middleware.GetUserID(c)
	action, svcErr := h.svc.CreateEnforcement(userID, req, c.IP())
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, action)
}

func (h *Handler) ListEnforcementsByUser(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("userId"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid user id"))
	}

	items, total, svcErr := h.svc.ListEnforcementsByUser(userID, 0, 100)
	if svcErr != nil {
		if appErr, ok := svcErr.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{
		"items": items,
		"total": total,
	})
}

// --- Audit Logs ---

func (h *Handler) ListAuditLogs(c *fiber.Ctx) error {
	var pg types.PaginationRequest
	if err := c.QueryParser(&pg); err != nil {
		return response.Err(c, errors.ErrBadRequest.WithMessage("invalid pagination params"))
	}
	pg.Normalize()

	var actorID uint64
	if actorIDStr := c.Query("actor_id"); actorIDStr != "" {
		parsed, err := strconv.ParseUint(actorIDStr, 10, 64)
		if err != nil {
			return response.Err(c, errors.ErrBadRequest.WithMessage("invalid actor_id"))
		}
		actorID = parsed
	}
	action := c.Query("action")

	logs, total, err := h.svc.ListAuditLogs(actorID, action, pg.Offset(), pg.PageSize)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Paginated(c, logs, types.NewPaginationMeta(pg, total))
}
