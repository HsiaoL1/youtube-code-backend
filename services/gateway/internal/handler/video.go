package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
	"youtube-code-backend/services/gateway/internal/viewmodel"
)

type VideoHandler struct {
	videoRepo *repository.VideoRepo
	userRepo  *repository.UserRepo
}

func NewVideoHandler(videoRepo *repository.VideoRepo, userRepo *repository.UserRepo) *VideoHandler {
	return &VideoHandler{videoRepo: videoRepo, userRepo: userRepo}
}

func (h *VideoHandler) Detail(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	if id == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Not found"})
	}

	row, err := h.videoRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Not found"})
	}

	videos, err := enrichVideos([]repository.VideoDBRow{*row}, h.userRepo)
	if err != nil || len(videos) == 0 {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load video"})
	}

	// Hardcoded chapters (no chapters table in DB)
	chapters := []viewmodel.ChapterVM{
		{ID: "ch1", Title: "Introduction", Time: 0},
		{ID: "ch2", Title: "Data Layer", Time: 220},
		{ID: "ch3", Title: "Route Strategy", Time: 570},
		{ID: "ch4", Title: "Q&A", Time: 980},
	}

	return c.JSON(fiber.Map{
		"video":    videos[0],
		"chapters": chapters,
	})
}

func (h *VideoHandler) Recommendations(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))

	// Get the current video to know its category
	current, _ := h.videoRepo.FindByID(id)
	category := ""
	if current != nil {
		category = current.Category
	}

	// Get videos from same category first, then others
	rows, err := h.videoRepo.FindAll(category, "views", "", "ready", "public", 0, 20)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load recommendations"})
	}

	// Filter out the current video
	filtered := make([]repository.VideoDBRow, 0, len(rows))
	for _, r := range rows {
		if r.ID != id {
			filtered = append(filtered, r)
		}
	}

	// If not enough, add from other categories
	if len(filtered) < 10 {
		others, _ := h.videoRepo.FindAll("", "views", "", "ready", "public", 0, 20)
		for _, r := range others {
			if r.ID != id {
				dup := false
				for _, f := range filtered {
					if f.ID == r.ID {
						dup = true
						break
					}
				}
				if !dup {
					filtered = append(filtered, r)
				}
			}
		}
	}

	items, err := enrichVideos(filtered, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich videos"})
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *VideoHandler) Like(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	userID := middleware.GetUserID(c)
	if userID == 0 {
		userID = 1 // fallback for unauthenticated
	}
	newCount, err := h.videoRepo.IncrementLike(id, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to like"})
	}
	return c.JSON(fiber.Map{"likes": newCount})
}

func (h *VideoHandler) Favorite(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	userID := middleware.GetUserID(c)
	if userID == 0 {
		userID = 1
	}
	_ = h.videoRepo.CreateFavorite(id, userID)
	return c.JSON(fiber.Map{"ok": true})
}
