package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
	"youtube-code-backend/services/gateway/internal/viewmodel"
)

type AdminHandler struct {
	videoRepo  *repository.VideoRepo
	reportRepo *repository.ReportRepo
	userRepo   *repository.UserRepo
}

func NewAdminHandler(videoRepo *repository.VideoRepo, reportRepo *repository.ReportRepo, userRepo *repository.UserRepo) *AdminHandler {
	return &AdminHandler{videoRepo: videoRepo, reportRepo: reportRepo, userRepo: userRepo}
}

func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	pendingCount, _ := h.videoRepo.CountByStatus("processing")
	reportsCount, _ := h.reportRepo.CountOpen()
	userRows, _ := h.userRepo.ListAll()

	// Count creators (role != "user")
	creatorCount := 0
	for _, u := range userRows {
		if u.Role != "user" {
			creatorCount++
		}
	}

	stats := []fiber.Map{
		{"label": "Pending Reviews", "value": pendingCount},
		{"label": "Reports Open", "value": reportsCount},
		{"label": "Active Creators", "value": creatorCount},
	}
	return c.JSON(fiber.Map{"stats": stats})
}

func (h *AdminHandler) ReviewQueue(c *fiber.Ctx) error {
	rows, err := h.videoRepo.FindByStatusIn([]string{"processing", "rejected"})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load review queue"})
	}
	items, err := enrichVideos(rows, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich videos"})
	}
	return c.JSON(fiber.Map{"items": items})
}

type reviewActionRequest struct {
	Action string `json:"action"` // "approve"|"reject"|"take_down"
}

func (h *AdminHandler) ReviewAction(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))

	var req reviewActionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}

	var dbStatus string
	switch req.Action {
	case "approve":
		dbStatus = "ready"
	case "reject":
		dbStatus = "rejected"
	case "take_down":
		dbStatus = "draft"
	default:
		return c.Status(400).JSON(fiber.Map{"message": "Invalid action"})
	}

	fields := map[string]any{"status": dbStatus}
	if err := h.videoRepo.Update(id, fields); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to update video"})
	}

	row, _ := h.videoRepo.FindByID(id)
	if row == nil {
		return c.JSON(fiber.Map{"ok": true, "item": nil})
	}
	items, _ := enrichVideos([]repository.VideoDBRow{*row}, h.userRepo)
	var item *viewmodel.VideoVM
	if len(items) > 0 {
		item = &items[0]
	}
	return c.JSON(fiber.Map{"ok": true, "item": item})
}

func (h *AdminHandler) Reports(c *fiber.Ctx) error {
	rows, err := h.reportRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load reports"})
	}

	items := make([]fiber.Map, 0, len(rows))
	for _, r := range rows {
		items = append(items, fiber.Map{
			"id":     convert.Uint64ToStr(r.ID),
			"reason": r.Reason,
			"target": convert.Uint64ToStr(r.ContentID),
			"status": r.Status,
		})
	}
	return c.JSON(fiber.Map{"items": items})
}

type reportActionRequest struct {
	Action string `json:"action"`
}

func (h *AdminHandler) ReportAction(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))

	var req reportActionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}

	status := "resolved"
	if req.Action == "dismiss" {
		status = "dismissed"
	}
	_ = h.reportRepo.UpdateStatus(id, status)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *AdminHandler) Users(c *fiber.Ctx) error {
	rows, err := h.userRepo.ListAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load users"})
	}

	items := make([]fiber.Map, 0, len(rows))
	for _, r := range rows {
		user := convert.UserRowToVM(r)
		items = append(items, fiber.Map{
			"id":             user.ID,
			"name":           user.Name,
			"handle":         user.Handle,
			"avatarUrl":      user.AvatarURL,
			"role":           user.Role,
			"followersCount": user.FollowersCount,
			"banned":         r.Status == "banned",
		})
	}
	return c.JSON(fiber.Map{"items": items})
}

type userActionRequest struct {
	Action string `json:"action"` // "ban"|"unban"|"warn"
}

func (h *AdminHandler) UserAction(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))

	var req userActionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}

	switch req.Action {
	case "ban":
		_ = h.userRepo.UpdateUserStatus(id, "banned")
	case "unban":
		_ = h.userRepo.UpdateUserStatus(id, "active")
	}
	return c.JSON(fiber.Map{"ok": true})
}
