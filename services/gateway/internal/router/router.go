package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/gateway/internal/handler"
	"youtube-code-backend/services/gateway/internal/repository"
)

func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Repositories
	userRepo := repository.NewUserRepo(db)
	videoRepo := repository.NewVideoRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	liveRepo := repository.NewLiveRepo(db)
	chatRepo := repository.NewChatRepo(db)
	subRepo := repository.NewSubscriptionRepo(db)
	playlistRepo := repository.NewPlaylistRepo(db)
	reportRepo := repository.NewReportRepo(db)

	// Handlers
	authH := handler.NewAuthHandler(userRepo, jm)
	feedH := handler.NewFeedHandler(videoRepo, userRepo, subRepo)
	searchH := handler.NewSearchHandler(videoRepo, liveRepo, userRepo)
	videoH := handler.NewVideoHandler(videoRepo, userRepo)
	shortsH := handler.NewShortsHandler(videoRepo, userRepo)
	liveH := handler.NewLiveHandler(liveRepo, chatRepo, userRepo, videoRepo)
	commentH := handler.NewCommentHandler(commentRepo, userRepo)
	channelH := handler.NewChannelHandler(userRepo, videoRepo)
	playlistH := handler.NewPlaylistHandler(playlistRepo, videoRepo, userRepo)
	studioH := handler.NewStudioHandler(videoRepo, commentRepo, userRepo, liveRepo, subRepo)
	adminH := handler.NewAdminHandler(videoRepo, reportRepo, userRepo)

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login", authH.Login)
	auth.Post("/register", authH.Register)
	auth.Get("/me", middleware.OptionalJWTAuth(jm), authH.Me)
	auth.Post("/logout", authH.Logout)

	// Feed routes
	feed := api.Group("/feed")
	feed.Get("/home", feedH.Home)
	feed.Get("/trending", feedH.Trending)
	feed.Get("/subscriptions", middleware.OptionalJWTAuth(jm), feedH.Subscriptions)

	// Search
	api.Get("/search", searchH.Search)

	// Video routes
	videos := api.Group("/videos")
	videos.Get("/:id", videoH.Detail)
	videos.Get("/:id/recommendations", videoH.Recommendations)
	videos.Post("/:id/like", middleware.OptionalJWTAuth(jm), videoH.Like)
	videos.Post("/:id/favorite", middleware.OptionalJWTAuth(jm), videoH.Favorite)

	// Shorts routes
	shorts := api.Group("/shorts")
	shorts.Get("/", shortsH.List)
	shorts.Post("/:id/like", middleware.OptionalJWTAuth(jm), shortsH.Like)
	shorts.Post("/:id/favorite", middleware.OptionalJWTAuth(jm), shortsH.Favorite)

	// Live routes
	live := api.Group("/live")
	live.Get("/", liveH.List)
	live.Get("/:id", liveH.Detail)
	live.Get("/:id/chat", liveH.Chat)

	// Comment routes
	comments := api.Group("/comments")
	comments.Get("/", commentH.List)
	comments.Post("/", middleware.JWTAuth(jm), commentH.Create)
	comments.Delete("/:id", middleware.JWTAuth(jm), commentH.Delete)
	comments.Post("/:id/like", middleware.OptionalJWTAuth(jm), commentH.Like)

	// Channel routes
	api.Get("/channel/:id", channelH.Detail)

	// Playlist routes
	api.Get("/playlists/:id", playlistH.Detail)

	// Studio routes (protected)
	studio := api.Group("/studio", middleware.JWTAuth(jm))
	studio.Get("/overview", studioH.Overview)
	studio.Get("/content", studioH.Content)
	studio.Post("/upload", studioH.Upload)
	studio.Patch("/content/:id", studioH.Update)
	studio.Post("/live/toggle", studioH.ToggleLive)
	studio.Get("/comments", studioH.Comments)

	// Admin routes (protected, admin only)
	admin := api.Group("/admin", middleware.JWTAuth(jm), middleware.RequireRole("admin"))
	admin.Get("/dashboard", adminH.Dashboard)
	admin.Get("/review-queue", adminH.ReviewQueue)
	admin.Post("/review-queue/:id/action", adminH.ReviewAction)
	admin.Get("/reports", adminH.Reports)
	admin.Post("/reports/:id/action", adminH.ReportAction)
	admin.Get("/users", adminH.Users)
	admin.Post("/users/:id/action", adminH.UserAction)
}
