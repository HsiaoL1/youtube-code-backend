package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
)

type CommentHandler struct {
	commentRepo *repository.CommentRepo
	userRepo    *repository.UserRepo
}

func NewCommentHandler(commentRepo *repository.CommentRepo, userRepo *repository.UserRepo) *CommentHandler {
	return &CommentHandler{commentRepo: commentRepo, userRepo: userRepo}
}

func (h *CommentHandler) List(c *fiber.Ctx) error {
	entityType := c.Query("entityType", "video")
	entityID := convert.StrToUint64(c.Query("entityId"))
	if entityID == 0 {
		return c.JSON(fiber.Map{"items": []any{}})
	}

	rows, err := h.commentRepo.FindByVideoID(entityID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load comments"})
	}
	items, err := enrichComments(rows, h.userRepo, entityType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich comments"})
	}
	return c.JSON(fiber.Map{"items": items})
}

type createCommentRequest struct {
	EntityType string  `json:"entityType"`
	EntityID   string  `json:"entityId"`
	Content    string  `json:"content"`
	ParentID   *string `json:"parentId,omitempty"`
}

func (h *CommentHandler) Create(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return c.Status(401).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var req createCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}

	videoID := convert.StrToUint64(req.EntityID)
	var parentID *uint64
	if req.ParentID != nil {
		pid := convert.StrToUint64(*req.ParentID)
		if pid > 0 {
			parentID = &pid
		}
	}

	id, err := h.commentRepo.Create(videoID, userID, req.Content, parentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to create comment"})
	}

	row, err := h.commentRepo.FindByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load comment"})
	}

	entityType := req.EntityType
	if entityType == "" {
		entityType = "video"
	}

	items, err := enrichComments([]repository.CommentDBRow{*row}, h.userRepo, entityType)
	if err != nil || len(items) == 0 {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich comment"})
	}

	return c.Status(201).JSON(fiber.Map{"item": items[0]})
}

func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	if err := h.commentRepo.SoftDelete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to delete comment"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *CommentHandler) Like(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	userID := middleware.GetUserID(c)
	if userID == 0 {
		userID = 1
	}
	newCount, err := h.commentRepo.IncrementLike(id, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to like comment"})
	}
	return c.JSON(fiber.Map{"likeCount": newCount})
}
