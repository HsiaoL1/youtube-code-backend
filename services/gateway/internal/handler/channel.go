package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
)

type ChannelHandler struct {
	userRepo  *repository.UserRepo
	videoRepo *repository.VideoRepo
}

func NewChannelHandler(userRepo *repository.UserRepo, videoRepo *repository.VideoRepo) *ChannelHandler {
	return &ChannelHandler{userRepo: userRepo, videoRepo: videoRepo}
}

func (h *ChannelHandler) Detail(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	if id == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Not found"})
	}

	row, err := h.userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Not found"})
	}
	channel := convert.UserRowToVM(*row)

	// Get the user's channel ID for video lookup
	channelID, _ := h.userRepo.FindChannelByUserID(id)
	var content []any
	if channelID > 0 {
		videoRows, err := h.videoRepo.FindAll("", "latest", "", "ready", "public", channelID, 50)
		if err == nil {
			items, err := enrichVideos(videoRows, h.userRepo)
			if err == nil {
				for _, item := range items {
					content = append(content, item)
				}
			}
		}
	}
	if content == nil {
		content = []any{}
	}

	return c.JSON(fiber.Map{
		"channel": channel,
		"content": content,
	})
}
