package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/chat/internal/handler"
	"youtube-code-backend/services/chat/internal/model"
	"youtube-code-backend/services/chat/internal/repository"
	"youtube-code-backend/services/chat/internal/service"
)

// Register wires service routes under /api/v1.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate models
	database.AutoMigrate(db,
		&model.ChatMessage{},
		&model.ChatRoomConfig{},
		&model.ChatManager{},
		&model.ChatMute{},
		&model.ChatKeywordFilter{},
	)

	// Build dependency chain
	messageRepo := repository.NewMessageRepository(db)
	configRepo := repository.NewRoomConfigRepository(db)
	managerRepo := repository.NewManagerRepository(db)
	muteRepo := repository.NewMuteRepository(db)
	filterRepo := repository.NewKeywordFilterRepository(db)

	svc := service.NewChatService(messageRepo, configRepo, managerRepo, muteRepo, filterRepo)
	h := handler.New(svc)

	// Auth middleware
	auth := middleware.JWTAuth(jm)

	// Chat routes
	chat := api.Group("/chat")

	// TODO: WebSocket upgrade at /ws/:roomId

	// Messages
	chat.Get("/messages/:roomId", h.GetMessages)
	chat.Post("/rooms/:roomId/send", auth, h.SendMessage)

	// Room config
	chat.Get("/rooms/:roomId/config", h.GetRoomConfig)
	chat.Put("/rooms/:roomId/config", auth, h.UpdateRoomConfig)

	// Managers
	chat.Get("/rooms/:roomId/managers", h.ListManagers)
	chat.Post("/rooms/:roomId/managers", auth, h.AddManager)
	chat.Delete("/rooms/:roomId/managers/:userId", auth, h.RemoveManager)

	// Mutes
	chat.Post("/rooms/:roomId/mute", auth, h.MuteUser)
	chat.Delete("/rooms/:roomId/mute/:userId", auth, h.UnmuteUser)

	// Keyword filters
	chat.Get("/rooms/:roomId/filters", h.ListFilters)
	chat.Post("/rooms/:roomId/filters", auth, h.AddFilter)
	chat.Delete("/rooms/:roomId/filters/:id", auth, h.RemoveFilter)
}
