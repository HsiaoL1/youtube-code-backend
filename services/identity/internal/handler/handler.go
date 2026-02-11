package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/pkg/common/response"
	"youtube-code-backend/services/identity/internal/model"
	"youtube-code-backend/services/identity/internal/service"
)

type Handler struct {
	authService *service.AuthService
}

func New(authService *service.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req service.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	user, tokens, err := h.authService.Register(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.Created(c, fiber.Map{"user": user, "tokens": tokens})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req service.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	user, tokens, err := h.authService.Login(req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"user": user, "tokens": tokens})
}

func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.RefreshToken == "" {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	tokens, err := h.authService.RefreshToken(body.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, tokens)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.RefreshToken == "" {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	if err := h.authService.Logout(body.RefreshToken); err != nil {
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) LogoutAll(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if err := h.authService.LogoutAll(userID); err != nil {
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) GetMe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	user, err := h.authService.GetMe(userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, user)
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req service.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	if err := h.authService.ChangePassword(userID, req); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.NoContent(c)
}

func (h *Handler) RequestPasswordReset(c *fiber.Ctx) error {
	var req service.PasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	_ = h.authService.RequestPasswordReset(req)
	return response.OK(c, fiber.Map{"message": "if the email exists, a reset code has been sent"})
}

func (h *Handler) ConfirmPasswordReset(c *fiber.Ctx) error {
	var req service.PasswordResetConfirm
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	if err := h.authService.ConfirmPasswordReset(req); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "password has been reset"})
}

func (h *Handler) SendVerification(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req service.VerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	if err := h.authService.SendVerification(userID, req); err != nil {
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "verification code sent"})
}

func (h *Handler) ConfirmVerification(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req service.VerificationConfirm
	if err := c.BodyParser(&req); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	if err := h.authService.ConfirmVerification(userID, req); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return response.Err(c, appErr)
		}
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "verified"})
}

func (h *Handler) UpdateUserRole(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest)
	}
	var body struct {
		Role model.UserRole `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	if err := h.authService.UpdateUserRole(id, body.Role); err != nil {
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "role updated"})
}

func (h *Handler) UpdateUserStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return response.Err(c, errors.ErrBadRequest)
	}
	var body struct {
		Status model.UserStatus `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.Err(c, errors.ErrInvalidPayload)
	}
	if err := h.authService.UpdateUserStatus(id, body.Status); err != nil {
		return response.Err(c, errors.ErrInternal)
	}
	return response.OK(c, fiber.Map{"message": "status updated"})
}
