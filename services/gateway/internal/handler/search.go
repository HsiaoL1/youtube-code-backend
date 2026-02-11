package handler

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/services/gateway/internal/convert"
	"youtube-code-backend/services/gateway/internal/repository"
	"youtube-code-backend/services/gateway/internal/viewmodel"
)

type SearchHandler struct {
	videoRepo *repository.VideoRepo
	liveRepo  *repository.LiveRepo
	userRepo  *repository.UserRepo
}

func NewSearchHandler(videoRepo *repository.VideoRepo, liveRepo *repository.LiveRepo, userRepo *repository.UserRepo) *SearchHandler {
	return &SearchHandler{videoRepo: videoRepo, liveRepo: liveRepo, userRepo: userRepo}
}

func (h *SearchHandler) Search(c *fiber.Ctx) error {
	q := c.Query("q", "")
	tab := c.Query("tab", "all")
	sort := c.Query("sort", "")

	// Search videos (long + replay)
	videoRows, err := h.videoRepo.SearchByTitle(q, "video")
	if err != nil {
		videoRows = nil
	}
	videos, _ := enrichVideos(videoRows, h.userRepo)
	if sort == "latest" || sort == "views" {
		// Already sorted by the query
	}

	// Search shorts
	shortRows, err := h.videoRepo.SearchByTitle(q, "short")
	if err != nil {
		shortRows = nil
	}
	shorts, _ := enrichVideos(shortRows, h.userRepo)

	// Search live rooms
	liveRows, err := h.liveRepo.SearchByTitle(q)
	if err != nil {
		liveRows = nil
	}
	lives, _ := enrichLiveRooms(liveRows, h.userRepo)

	// Search channels (users by name)
	allUsers, err := h.userRepo.ListAll()
	if err != nil {
		allUsers = nil
	}
	channels := []viewmodel.UserVM{}
	qLower := toLower(q)
	for _, u := range allUsers {
		name := u.Nickname
		if name == "" {
			name = u.Username
		}
		if q == "" || containsLower(name, qLower) || containsLower(u.Username, qLower) {
			channels = append(channels, convert.UserRowToVM(u))
		}
	}

	return c.JSON(fiber.Map{
		"videos":    videos,
		"shorts":    shorts,
		"lives":     lives,
		"channels":  channels,
		"activeTab": tab,
	})
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, sub string) bool {
	s = toLower(s)
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
