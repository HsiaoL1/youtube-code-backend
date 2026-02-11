package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/notification/internal/handler"
	"youtube-code-backend/services/notification/internal/model"
	"youtube-code-backend/services/notification/internal/repository"
	"youtube-code-backend/services/notification/internal/service"
)

// Register wires service routes under /api/v1.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate all models
	database.AutoMigrate(db,
		&model.Notification{},
		&model.NotificationPreference{},
	)

	// Repositories
	notifRepo := repository.NewNotificationRepository(db)
	prefRepo := repository.NewPreferenceRepository(db)

	// Services
	notifService := service.NewNotificationService(notifRepo, prefRepo)

	h := handler.New(notifService)

	auth := middleware.JWTAuth(jm)

	// --- Notification routes: /api/v1/notifications ---
	notifications := api.Group("/notifications")

	// Authenticated routes
	notifAuth := notifications.Group("", auth)
	notifAuth.Get("/", h.ListNotifications)
	notifAuth.Get("/unread/count", h.UnreadCount)
	notifAuth.Put("/read-all", h.MarkAllAsRead)
	notifAuth.Put("/:id/read", h.MarkAsRead)
	notifAuth.Delete("/:id", h.DeleteNotification)
	notifAuth.Get("/preferences", h.GetPreferences)
	notifAuth.Put("/preferences", h.UpdatePreferences)

	// Internal routes (no auth — called by other services)
	notifications.Post("/send", h.SendNotification)
	notifications.Post("/broadcast", h.BroadcastNotification)
}
