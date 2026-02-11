package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/user-channel/internal/handler"
	"youtube-code-backend/services/user-channel/internal/model"
	"youtube-code-backend/services/user-channel/internal/repository"
	"youtube-code-backend/services/user-channel/internal/service"
)

// Register wires service routes under /api/v1/user-channel.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate models
	database.AutoMigrate(db, &model.UserProfile{}, &model.Channel{}, &model.ChannelLink{})

	// Repositories
	profileRepo := repository.NewProfileRepository(db)
	channelRepo := repository.NewChannelRepository(db)

	// Services
	profileService := service.NewProfileService(profileRepo)
	channelService := service.NewChannelService(channelRepo)

	h := handler.New(profileService, channelService)

	g := api.Group("/user-channel")

	// Public routes
	g.Get("/profiles/:userId", h.GetProfile)
	g.Get("/channels/:id", h.GetChannel)
	g.Get("/channels/handle/:handle", h.GetChannelByHandle)
	g.Get("/channels/:id/stats", h.GetChannelStats)

	// Authenticated routes
	auth := g.Group("", middleware.JWTAuth(jm))
	auth.Put("/profiles", h.UpdateProfile)
	auth.Post("/channels", h.CreateChannel)
	auth.Put("/channels/:id", h.UpdateChannel)
}
