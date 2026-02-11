package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
)

type ShortsHandler struct {
	videoRepo *repository.VideoRepo
	userRepo  *repository.UserRepo
}

func NewShortsHandler(videoRepo *repository.VideoRepo, userRepo *repository.UserRepo) *ShortsHandler {
	return &ShortsHandler{videoRepo: videoRepo, userRepo: userRepo}
}

func (h *ShortsHandler) List(c *fiber.Ctx) error {
	rows, err := h.videoRepo.FindAll("", "latest", "short", "ready", "public", 0, 50)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load shorts"})
	}
	items, err := enrichVideos(rows, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich shorts"})
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *ShortsHandler) Like(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	userID := middleware.GetUserID(c)
	if userID == 0 {
		userID = 1
	}
	newCount, err := h.videoRepo.IncrementLike(id, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to like"})
	}
	return c.JSON(fiber.Map{"likes": newCount})
}

func (h *ShortsHandler) Favorite(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	userID := middleware.GetUserID(c)
	if userID == 0 {
		userID = 1
	}
	_ = h.videoRepo.CreateFavorite(id, userID)
	return c.JSON(fiber.Map{"ok": true})
}
