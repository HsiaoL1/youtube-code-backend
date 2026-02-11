package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/gateway/internal/repository"
)

type FeedHandler struct {
	videoRepo *repository.VideoRepo
	userRepo  *repository.UserRepo
	subRepo   *repository.SubscriptionRepo
}

func NewFeedHandler(videoRepo *repository.VideoRepo, userRepo *repository.UserRepo, subRepo *repository.SubscriptionRepo) *FeedHandler {
	return &FeedHandler{videoRepo: videoRepo, userRepo: userRepo, subRepo: subRepo}
}

func (h *FeedHandler) Home(c *fiber.Ctx) error {
	category := c.Query("category", "All")
	rows, err := h.videoRepo.FindAll(category, "latest", "", "ready", "public", 0, 50)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load feed"})
	}
	items, err := enrichVideos(rows, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich videos"})
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *FeedHandler) Trending(c *fiber.Ctx) error {
	rows, err := h.videoRepo.FindAll("", "views", "", "ready", "public", 0, 50)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load trending"})
	}
	items, err := enrichVideos(rows, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich videos"})
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *FeedHandler) Subscriptions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.JSON(fiber.Map{"items": []any{}})
	}

	channelIDs, err := h.subRepo.FindChannelIDsBySubscriber(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load subscriptions"})
	}
	if len(channelIDs) == 0 {
		return c.JSON(fiber.Map{"items": []any{}})
	}

	rows, err := h.videoRepo.FindByChannelIDs(channelIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load videos"})
	}
	items, err := enrichVideos(rows, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich videos"})
	}
	return c.JSON(fiber.Map{"items": items})
}
