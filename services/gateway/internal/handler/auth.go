package handler

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
)

type AuthHandler struct {
	userRepo *repository.UserRepo
	jm       *jwt.Manager
}

func NewAuthHandler(userRepo *repository.UserRepo, jm *jwt.Manager) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jm: jm}
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}

	// Try username first, then email
	row, err := h.userRepo.FindByUsername(req.Identifier)
	if err != nil {
		row, err = h.userRepo.FindByEmail(req.Identifier)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"message": "Invalid credentials"})
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "Invalid credentials"})
	}

	pair, err := h.jm.GenerateTokenPair(row.UserID, row.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to generate token"})
	}

	user := convert.UserRowToVM(*row)
	return c.JSON(fiber.Map{
		"user":  user,
		"token": pair.AccessToken,
	})
}

type registerRequest struct {
	Name     string `json:"name"`
	Handle   string `json:"handle"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to hash password"})
	}

	// Use handle as username, generate email
	handle := req.Handle
	if len(handle) > 0 && handle[0] == '@' {
		handle = handle[1:]
	}
	email := handle + "@example.com"
	avatar := "https://i.pravatar.cc/120?img=65"

	userID, err := h.userRepo.CreateUser(handle, email, string(hash), req.Name, handle, avatar)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to create user"})
	}

	pair, err := h.jm.GenerateTokenPair(userID, "user")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to generate token"})
	}

	row, err := h.userRepo.FindByID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load user"})
	}

	user := convert.UserRowToVM(*row)
	return c.Status(201).JSON(fiber.Map{
		"user":  user,
		"token": pair.AccessToken,
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(fiber.Map{"user": nil})
	}
	row, err := h.userRepo.FindByID(userID)
	if err != nil {
		return c.JSON(fiber.Map{"user": nil})
	}
	user := convert.UserRowToVM(*row)
	return c.JSON(fiber.Map{"user": user})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ok": true})
}
