package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/social-graph/internal/handler"
	"youtube-code-backend/services/social-graph/internal/model"
	"youtube-code-backend/services/social-graph/internal/repository"
	"youtube-code-backend/services/social-graph/internal/service"
)

// Register wires service routes under /api/v1.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate models
	database.AutoMigrate(db, &model.Subscription{})

	// Build dependency chain
	repo := repository.NewSubscriptionRepository(db)
	svc := service.NewSubscriptionService(repo)
	h := handler.New(svc)

	// Auth middleware
	auth := middleware.JWTAuth(jm)

	// Social graph routes
	social := api.Group("/social")

	// Authenticated routes
	social.Post("/subscribe", auth, h.Subscribe)
	social.Delete("/subscribe/:channelId", auth, h.Unsubscribe)
	social.Get("/subscriptions", auth, h.ListSubscriptions)
	social.Get("/subscriptions/check/:channelId", auth, h.CheckSubscription)
	social.Put("/subscribe/:channelId/notify", auth, h.UpdateNotifyPreference)

	// Public routes
	social.Get("/subscribers/:channelId", h.ListSubscribers)
	social.Get("/subscribers/:channelId/count", h.GetSubscriberCount)

	// Internal endpoint
	social.Get("/followers/:channelId/ids", h.GetFollowerIDs)
}
