package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/comment/internal/handler"
	"youtube-code-backend/services/comment/internal/model"
	"youtube-code-backend/services/comment/internal/repository"
	"youtube-code-backend/services/comment/internal/service"
)

// Register wires service routes under /api/v1/comments.
func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate models
	database.AutoMigrate(db, &model.Comment{}, &model.CommentLike{}, &model.SensitiveWord{})

	// Repositories
	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewCommentLikeRepository(db)
	sensitiveWordRepo := repository.NewSensitiveWordRepository(db)

	// Services
	commentService := service.NewCommentService(commentRepo, likeRepo, sensitiveWordRepo)
	sensitiveWordService := service.NewSensitiveWordService(sensitiveWordRepo)

	h := handler.New(commentService, sensitiveWordService)

	g := api.Group("/comments")

	// Public / optional-auth routes
	g.Get("/video/:videoId", middleware.OptionalJWTAuth(jm), h.ListByVideo)
	g.Get("/:id/replies", middleware.OptionalJWTAuth(jm), h.ListReplies)

	// Authenticated routes
	auth := g.Group("", middleware.JWTAuth(jm))
	auth.Post("/", h.CreateComment)
	auth.Delete("/:id", h.DeleteComment)
	auth.Post("/:id/like", h.LikeComment)
	auth.Delete("/:id/like", h.UnlikeComment)
	auth.Post("/:id/pin", h.PinComment)
	auth.Post("/:id/heart", h.HeartComment)
	auth.Post("/:id/report", h.ReportComment)

	// Admin routes
	admin := g.Group("/sensitive-words", middleware.JWTAuth(jm), middleware.RequireRole("admin"))
	admin.Get("/", h.ListSensitiveWords)
	admin.Post("/", h.AddSensitiveWord)
	admin.Delete("/:id", h.DeleteSensitiveWord)
}
