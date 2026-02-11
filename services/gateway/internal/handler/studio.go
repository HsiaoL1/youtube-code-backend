package handler

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
)

type StudioHandler struct {
	videoRepo   *repository.VideoRepo
	commentRepo *repository.CommentRepo
	userRepo    *repository.UserRepo
	liveRepo    *repository.LiveRepo
	subRepo     *repository.SubscriptionRepo
}

func NewStudioHandler(
	videoRepo *repository.VideoRepo,
	commentRepo *repository.CommentRepo,
	userRepo *repository.UserRepo,
	liveRepo *repository.LiveRepo,
	subRepo *repository.SubscriptionRepo,
) *StudioHandler {
	return &StudioHandler{
		videoRepo:   videoRepo,
		commentRepo: commentRepo,
		userRepo:    userRepo,
		liveRepo:    liveRepo,
		subRepo:     subRepo,
	}
}

func (h *StudioHandler) Overview(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	channelID, _ := h.userRepo.FindChannelByUserID(userID)

	totalViews, _ := h.videoRepo.SumViewsByChannelID(channelID)
	subCount, _ := h.subRepo.GetSubscriberCount(channelID)

	cards := []fiber.Map{
		{"label": "Views", "value": formatCount(totalViews)},
		{"label": "Followers", "value": formatCount(subCount)},
		{"label": "Revenue", "value": "$0"},
	}

	videoRows, _ := h.videoRepo.FindAll("", "latest", "", "", "", channelID, 3)
	latest, _ := enrichVideos(videoRows, h.userRepo)

	return c.JSON(fiber.Map{
		"cards":  cards,
		"latest": latest,
	})
}

func (h *StudioHandler) Content(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	channelID, _ := h.userRepo.FindChannelByUserID(userID)

	status := c.Query("status", "all")
	dbStatus := ""
	if status != "all" {
		dbStatus = convert.VideoStatusToDB(status)
	}

	rows, err := h.videoRepo.FindAll("", "latest", "", dbStatus, "", channelID, 50)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load content"})
	}
	items, err := enrichVideos(rows, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich videos"})
	}
	return c.JSON(fiber.Map{"items": items})
}

type uploadRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CoverURL    string   `json:"coverUrl"`
	Tags        []string `json:"tags"`
	Category    string   `json:"category"`
	Visibility  string   `json:"visibility"`
	Status      string   `json:"status"`
}

func (h *StudioHandler) Upload(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	channelID, _ := h.userRepo.FindChannelByUserID(userID)

	var req uploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}

	if req.Title == "" {
		req.Title = "Untitled"
	}
	if req.Visibility == "" {
		req.Visibility = "public"
	}
	if req.Category == "" {
		req.Category = "Tech"
	}
	if req.CoverURL == "" {
		req.CoverURL = "https://picsum.photos/seed/upload/640/360"
	}

	tagsJSON, _ := json.Marshal(req.Tags)

	id, err := h.videoRepo.Create(channelID, req.Title, req.Description, req.CoverURL, "video", req.Category, req.Visibility, string(tagsJSON), "")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to upload video"})
	}

	row, _ := h.videoRepo.FindByID(id)
	if row == nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load video"})
	}
	items, _ := enrichVideos([]repository.VideoDBRow{*row}, h.userRepo)
	if len(items) == 0 {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich video"})
	}
	return c.Status(201).JSON(fiber.Map{"item": items[0]})
}

func (h *StudioHandler) Update(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))

	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}

	fields := map[string]any{}
	if v, ok := body["title"]; ok {
		fields["title"] = v
	}
	if v, ok := body["description"]; ok {
		fields["description"] = v
	}
	if v, ok := body["visibility"]; ok {
		fields["visibility"] = v
	}
	if v, ok := body["status"]; ok {
		if s, ok := v.(string); ok {
			fields["status"] = convert.VideoStatusToDB(s)
		}
	}
	if v, ok := body["category"]; ok {
		fields["category"] = v
	}
	if v, ok := body["coverUrl"]; ok {
		fields["thumbnail_url"] = v
	}
	if v, ok := body["tags"]; ok {
		if tags, ok := v.([]any); ok {
			tagsJSON, _ := json.Marshal(tags)
			fields["tags"] = string(tagsJSON)
		}
	}

	if err := h.videoRepo.Update(id, fields); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to update video"})
	}

	row, _ := h.videoRepo.FindByID(id)
	if row == nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load video"})
	}
	items, _ := enrichVideos([]repository.VideoDBRow{*row}, h.userRepo)
	if len(items) == 0 {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich video"})
	}
	return c.JSON(fiber.Map{"item": items[0]})
}

func (h *StudioHandler) ToggleLive(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	channelID, _ := h.userRepo.FindChannelByUserID(userID)
	if err := h.liveRepo.ToggleStatus(channelID); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to toggle live"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *StudioHandler) Comments(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	channelID, _ := h.userRepo.FindChannelByUserID(userID)

	// Get all videos for this channel
	videoRows, _ := h.videoRepo.FindAll("", "latest", "", "", "", channelID, 100)
	videoIDs := make([]uint64, 0, len(videoRows))
	for _, r := range videoRows {
		videoIDs = append(videoIDs, r.ID)
	}

	if len(videoIDs) == 0 {
		return c.JSON(fiber.Map{"items": []any{}})
	}

	commentRows, err := h.commentRepo.FindByVideoIDs(videoIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load comments"})
	}
	items, err := enrichComments(commentRows, h.userRepo, "video")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich comments"})
	}
	return c.JSON(fiber.Map{"items": items})
}

func formatCount(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.0fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
