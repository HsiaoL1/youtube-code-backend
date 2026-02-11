package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/feed-reco/internal/handler"
	"youtube-code-backend/services/feed-reco/internal/model"
	"youtube-code-backend/services/feed-reco/internal/repository"
	"youtube-code-backend/services/feed-reco/internal/service"
)

// Register wires feed-reco service routes under /api/v1.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate feed-reco owned tables.
	database.AutoMigrate(db, &model.TrendingScore{}, &model.CategoryFeed{})

	repo := repository.NewFeedRepository(db)
	feedService := service.NewFeedService(repo)
	h := handler.New(feedService)

	g := api.Group("/feed")

	// Public / optional-auth routes
	g.Get("/home", middleware.OptionalJWTAuth(jm), h.HomeFeed)
	g.Get("/trending", h.Trending)
	g.Get("/category/:category", h.CategoryFeed)
	g.Get("/shorts", h.ShortsFeed)
	g.Get("/related/:videoId", h.RelatedVideos)

	// Authenticated routes
	g.Get("/subscriptions", middleware.JWTAuth(jm), h.SubscriptionFeed)
}
