package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/analytics/internal/handler"
	"youtube-code-backend/services/analytics/internal/model"
	"youtube-code-backend/services/analytics/internal/repository"
	"youtube-code-backend/services/analytics/internal/service"
)

// Register wires service routes under /api/v1.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate all models
	database.AutoMigrate(db,
		&model.AnalyticsEvent{},
		&model.VideoDailyStats{},
		&model.ChannelDailyStats{},
		&model.LiveSessionStats{},
		&model.CounterSnapshot{},
	)

	// Repositories
	eventRepo := repository.NewEventRepository(db)
	videoStatsRepo := repository.NewVideoStatsRepository(db)
	channelStatsRepo := repository.NewChannelStatsRepository(db)
	liveStatsRepo := repository.NewLiveStatsRepository(db)
	counterRepo := repository.NewCounterRepository(db)

	// Services
	analyticsService := service.NewAnalyticsService(
		eventRepo, videoStatsRepo, channelStatsRepo, liveStatsRepo, counterRepo,
	)

	h := handler.New(analyticsService)

	auth := middleware.JWTAuth(jm)
	optionalAuth := middleware.OptionalJWTAuth(jm)
	adminOnly := middleware.RequireRole("admin")

	group := api.Group("/analytics")

	// Liveness
	group.Get("/ping", h.Ping)

	// Event ingestion — optional auth (anonymous events allowed)
	group.Post("/events", optionalAuth, h.IngestEvent)

	// Video analytics — requires auth
	group.Get("/videos/:videoId", auth, h.GetVideoAnalytics)
	group.Get("/videos/:videoId/realtime", h.GetVideoRealtime)

	// Channel analytics — requires auth
	group.Get("/channels/:channelId", auth, h.GetChannelAnalytics)
	group.Get("/channels/:channelId/overview", h.GetChannelOverview)

	// Live session stats
	group.Get("/live/:sessionId", h.GetLiveSessionStats)

	// Internal endpoints
	group.Post("/counters/increment", auth, h.IncrementCounter)
	group.Post("/aggregate/daily", auth, adminOnly, h.TriggerDailyAggregation)
}
