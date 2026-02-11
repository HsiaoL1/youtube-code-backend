package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/media/internal/handler"
	"youtube-code-backend/services/media/internal/model"
	"youtube-code-backend/services/media/internal/repository"
	"youtube-code-backend/services/media/internal/service"
)

func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate models
	database.AutoMigrate(db,
		&model.UploadSession{},
		&model.UploadPart{},
		&model.MediaAsset{},
		&model.MediaVariant{},
		&model.MediaThumbnail{},
		&model.MediaSubtitle{},
	)

	uploadRepo := repository.NewUploadRepository(db)
	assetRepo := repository.NewAssetRepository(db)

	uploadService := service.NewUploadService(uploadRepo)
	assetService := service.NewAssetService(assetRepo)

	h := handler.New(uploadService, assetService)

	g := api.Group("/media")

	// Authenticated upload routes
	auth := g.Group("", middleware.JWTAuth(jm))
	auth.Post("/upload/init", h.InitUpload)
	auth.Post("/upload/:sessionId/part", h.UploadPart)
	auth.Post("/upload/:sessionId/complete", h.CompleteUpload)
	auth.Post("/upload/:sessionId/abort", h.AbortUpload)
	auth.Get("/upload/:sessionId", h.GetSessionStatus)

	// Public asset routes
	g.Get("/assets/video/:videoId", h.GetAssetsByVideoID)
	g.Get("/assets/:id/variants", h.GetVariantsByAssetID)

	// Thumbnail routes
	g.Get("/thumbnails/video/:videoId", h.GetThumbnailsByVideoID)
	auth.Post("/thumbnails/video/:videoId", h.AddThumbnail)

	// Subtitle routes
	g.Get("/subtitles/video/:videoId", h.GetSubtitlesByVideoID)
	auth.Post("/subtitles/video/:videoId", h.AddSubtitle)
	auth.Delete("/subtitles/:id", h.DeleteSubtitle)
}
