package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/identity/internal/handler"
	"youtube-code-backend/services/identity/internal/model"
	"youtube-code-backend/services/identity/internal/repository"
	"youtube-code-backend/services/identity/internal/service"
)

func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate models
	database.AutoMigrate(db, &model.User{}, &model.RefreshToken{}, &model.VerificationCode{})

	repo := repository.NewUserRepository(db)
	authService := service.NewAuthService(repo, jm)
	h := handler.New(authService)

	g := api.Group("/identity")

	// Public routes
	g.Post("/register", h.Register)
	g.Post("/login", h.Login)
	g.Post("/token/refresh", h.RefreshToken)
	g.Post("/password/reset/request", h.RequestPasswordReset)
	g.Post("/password/reset/confirm", h.ConfirmPasswordReset)

	// Authenticated routes
	auth := g.Group("", middleware.JWTAuth(jm))
	auth.Get("/me", h.GetMe)
	auth.Post("/logout", h.Logout)
	auth.Post("/logout/all", h.LogoutAll)
	auth.Post("/password/change", h.ChangePassword)
	auth.Post("/verification/send", h.SendVerification)
	auth.Post("/verification/confirm", h.ConfirmVerification)

	// Admin routes
	admin := auth.Group("", middleware.RequireRole("admin"))
	admin.Put("/users/:id/role", h.UpdateUserRole)
	admin.Put("/users/:id/status", h.UpdateUserStatus)
}
