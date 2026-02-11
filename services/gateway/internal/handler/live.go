package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
	"youtube-code-backend/services/gateway/internal/viewmodel"
)

type LiveHandler struct {
	liveRepo *repository.LiveRepo
	chatRepo *repository.ChatRepo
	userRepo *repository.UserRepo
	videoRepo *repository.VideoRepo
}

func NewLiveHandler(liveRepo *repository.LiveRepo, chatRepo *repository.ChatRepo, userRepo *repository.UserRepo, videoRepo *repository.VideoRepo) *LiveHandler {
	return &LiveHandler{liveRepo: liveRepo, chatRepo: chatRepo, userRepo: userRepo, videoRepo: videoRepo}
}

func (h *LiveHandler) List(c *fiber.Ctx) error {
	category := c.Query("category", "All")
	rows, err := h.liveRepo.FindLive(category)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load live rooms"})
	}
	items, err := enrichLiveRooms(rows, h.userRepo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to enrich live rooms"})
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *LiveHandler) Detail(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	if id == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Not found"})
	}

	row, err := h.liveRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Not found"})
	}

	rooms, err := enrichLiveRooms([]repository.LiveRoomDBRow{*row}, h.userRepo)
	if err != nil || len(rooms) == 0 {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load room"})
	}

	// Find a replay video for this channel
	var replay *viewmodel.VideoVM
	videoRows, _ := h.videoRepo.FindAll("", "latest", "video", "ready", "", row.ChannelID, 5)
	if len(videoRows) > 0 {
		enriched, _ := enrichVideos(videoRows[:1], h.userRepo)
		if len(enriched) > 0 {
			// Mark as replay type
			enriched[0].Type = "replay"
			replay = &enriched[0]
		}
	}

	result := fiber.Map{"room": rooms[0]}
	if replay != nil {
		result["replay"] = *replay
	} else {
		result["replay"] = nil
	}
	return c.JSON(result)
}

func (h *LiveHandler) Chat(c *fiber.Ctx) error {
	id := convert.StrToUint64(c.Params("id"))
	rows, err := h.chatRepo.FindByRoomID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to load chat"})
	}

	// Enrich with user info - convert to CommentVM format
	userIDs := make([]uint64, 0, len(rows))
	seen := map[uint64]bool{}
	for _, r := range rows {
		if !seen[r.UserID] {
			userIDs = append(userIDs, r.UserID)
			seen[r.UserID] = true
		}
	}
	users, _ := h.userRepo.BatchFindUserDetails(userIDs)

	items := make([]viewmodel.CommentVM, 0, len(rows))
	for _, r := range rows {
		author := viewmodel.UserVM{}
		if ur, ok := users[r.UserID]; ok {
			author = convert.UserRowToVM(ur)
		}
		items = append(items, viewmodel.CommentVM{
			ID:         convert.Uint64ToStr(r.ID),
			EntityType: "live",
			EntityID:   convert.Uint64ToStr(r.RoomID),
			Author:     author,
			Content:    r.Content,
			CreatedAt:  r.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
			LikeCount:  0,
		})
	}
	return c.JSON(fiber.Map{"items": items})
}
