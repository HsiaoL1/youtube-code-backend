package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
	"youtube-code-backend/services/gateway/internal/viewmodel"
)

type PlaylistHandler struct {
	playlistRepo *repository.PlaylistRepo
	videoRepo    *repository.VideoRepo
	userRepo     *repository.UserRepo
}

func NewPlaylistHandler(playlistRepo *repository.PlaylistRepo, videoRepo *repository.VideoRepo, userRepo *repository.UserRepo) *PlaylistHandler {
	return &PlaylistHandler{playlistRepo: playlistRepo, videoRepo: videoRepo, userRepo: userRepo}
}

func (h *PlaylistHandler) Detail(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	if id == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Not found"})
	}

	pl, err := h.playlistRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Not found"})
	}

	plItems, err := h.playlistRepo.FindItems(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load playlist items"})
	}

	videoIDs := make([]uint64, 0, len(plItems))
	for _, item := range plItems {
		videoIDs = append(videoIDs, item.VideoID)
	}

	videoRows, err := h.videoRepo.FindByIDs(videoIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load videos"})
	}

	items, err := enrichVideos(videoRows, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich videos"})
	}

	// Reorder by playlist position
	videoMap := map[string]viewmodel.VideoVM{}
	for _, v := range items {
		videoMap[v.ID] = v
	}
	ordered := make([]viewmodel.VideoVM, 0, len(plItems))
	for _, pi := range plItems {
		vid := convert.Uint64ToStr(pi.VideoID)
		if v, ok := videoMap[vid]; ok {
			ordered = append(ordered, v)
		}
	}

	return c.JSON(fiber.Map{
		"playlist": viewmodel.PlaylistVM{
			ID:    convert.Uint64ToStr(pl.ID),
			Title: pl.Title,
			Items: ordered,
		},
	})
}
