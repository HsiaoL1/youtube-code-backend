package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/live/internal/handler"
	"youtube-code-backend/services/live/internal/model"
	"youtube-code-backend/services/live/internal/repository"
	"youtube-code-backend/services/live/internal/service"
)

// Register wires service routes under /api/v1/live.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate models
	database.AutoMigrate(db, &model.LiveRoom{}, &model.LiveSession{})

	roomRepo := repository.NewRoomRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	liveService := service.NewLiveService(roomRepo, sessionRepo)
	h := handler.New(liveService)

	g := api.Group("/live")

	// Public routes
	g.Get("/rooms/live", h.ListLiveRooms)
	g.Get("/rooms/channel/:channelId", h.ListChannelRooms)
	g.Get("/rooms/:id", h.GetRoom)
	g.Get("/rooms/:id/playback", h.GetPlaybackInfo)
	g.Get("/rooms/:id/sessions", h.ListSessions)

	// Internal route (stream auth callback)
	g.Post("/stream/auth", h.StreamAuth)

	// Authenticated routes
	auth := g.Group("", middleware.JWTAuth(jm))
	auth.Post("/rooms", h.CreateRoom)
	auth.Put("/rooms/:id", h.UpdateRoom)
	auth.Delete("/rooms/:id", h.DeleteRoom)
	auth.Post("/rooms/:id/stream-key/regenerate", h.RegenerateStreamKey)
	auth.Post("/rooms/:id/go-live", h.GoLive)
	auth.Post("/rooms/:id/end", h.EndStream)
}
