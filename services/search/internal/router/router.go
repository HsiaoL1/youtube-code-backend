package router

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/services/search/internal/handler"
	"youtube-code-backend/services/search/internal/model"
	"youtube-code-backend/services/search/internal/repository"
	"youtube-code-backend/services/search/internal/service"
)

// Register wires service routes under /api/v1.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate models
	database.AutoMigrate(db,
		&model.SearchVideo{},
		&model.SearchChannel{},
		&model.SearchLive{},
	)

	// Create GIN indexes for full-text search
	createGINIndexes(db)

	// Build dependency chain
	videoRepo := repository.NewVideoRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	liveRepo := repository.NewLiveRepository(db)
	svc := service.NewSearchService(videoRepo, channelRepo, liveRepo)
	h := handler.New(svc)

	// Search routes
	search := api.Group("/search")

	// Public search endpoints
	search.Get("/videos", h.SearchVideos)
	search.Get("/channels", h.SearchChannels)
	search.Get("/live", h.SearchLive)
	search.Get("/all", h.SearchAll)

	// Internal index management endpoints
	search.Post("/index/videos", h.IndexVideo)
	search.Delete("/index/videos/:videoId", h.DeleteVideoIndex)
	search.Post("/index/channels", h.IndexChannel)
	search.Post("/index/live", h.IndexLive)
}

// createGINIndexes creates PostgreSQL GIN indexes for full-text search.
func createGINIndexes(db *gorm.DB) {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_search_videos_tsv ON search_videos USING GIN (to_tsvector('english', title || ' ' || coalesce(description, '')))`,
		`CREATE INDEX IF NOT EXISTS idx_search_channels_tsv ON search_channels USING GIN (to_tsvector('english', name || ' ' || coalesce(handle, '') || ' ' || coalesce(description, '')))`,
		`CREATE INDEX IF NOT EXISTS idx_search_live_tsv ON search_live USING GIN (to_tsvector('english', title || ' ' || coalesce(channel_name, '')))`,
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("warning: failed to create GIN index: %v", err)
		}
	}
}
