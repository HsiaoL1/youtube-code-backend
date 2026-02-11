package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/video/internal/handler"
	"youtube-code-backend/services/video/internal/model"
	"youtube-code-backend/services/video/internal/repository"
	"youtube-code-backend/services/video/internal/service"
)

func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate all models
	database.AutoMigrate(db,
		&model.Video{},
		&model.VideoLike{},
		&model.VideoFavorite{},
		&model.WatchHistory{},
		&model.Playlist{},
		&model.PlaylistItem{},
	)

	// Repositories
	videoRepo := repository.NewVideoRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favRepo := repository.NewFavoriteRepository(db)
	watchRepo := repository.NewWatchRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)

	// Services
	videoService := service.NewVideoService(videoRepo, likeRepo, favRepo, watchRepo)
	playlistService := service.NewPlaylistService(playlistRepo, videoRepo)

	h := handler.New(videoService, playlistService)

	auth := middleware.JWTAuth(jm)

	// --- Video routes: /api/v1/videos ---
	videos := api.Group("/videos")

	// Static routes must be registered before parameterized routes.
	videos.Get("/history", auth, h.GetWatchHistory)

	videos.Get("/:id", h.GetVideo)

	videosAuth := videos.Group("", auth)
	videosAuth.Post("/", h.CreateVideo)
	videosAuth.Put("/:id", h.UpdateVideo)
	videosAuth.Delete("/:id", h.DeleteVideo)
	videosAuth.Post("/:id/publish", h.PublishVideo)
	videosAuth.Post("/:id/like", h.LikeVideo)
	videosAuth.Post("/:id/dislike", h.DislikeVideo)
	videosAuth.Delete("/:id/like", h.RemoveLike)
	videosAuth.Post("/:id/favorite", h.AddFavorite)
	videosAuth.Delete("/:id/favorite", h.RemoveFavorite)
	videosAuth.Get("/:id/progress", h.GetProgress)
	videosAuth.Post("/:id/progress", h.UpdateProgress)

	// --- Playlist routes: /api/v1/playlists ---
	playlists := api.Group("/playlists")

	playlists.Get("/:id", h.GetPlaylist)

	playlistsAuth := playlists.Group("", auth)
	playlistsAuth.Post("/", h.CreatePlaylist)
	playlistsAuth.Put("/:id", h.UpdatePlaylist)
	playlistsAuth.Delete("/:id", h.DeletePlaylist)
	playlistsAuth.Post("/:id/items", h.AddPlaylistItem)
	playlistsAuth.Delete("/:id/items/:videoId", h.RemovePlaylistItem)
}
