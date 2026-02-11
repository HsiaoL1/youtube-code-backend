package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/middleware"
	"youtube-code-backend/services/moderation/internal/handler"
	"youtube-code-backend/services/moderation/internal/model"
	"youtube-code-backend/services/moderation/internal/repository"
	"youtube-code-backend/services/moderation/internal/service"
)

func Register(api fiber.Router, db *gorm.DB, jm *jwt.Manager) {
	// Auto-migrate all models
	database.AutoMigrate(db,
		&model.ModerationItem{},
		&model.Report{},
		&model.EnforcementAction{},
		&model.AuditLog{},
	)

	// Repositories
	itemRepo := repository.NewModerationItemRepository(db)
	reportRepo := repository.NewReportRepository(db)
	enforcementRepo := repository.NewEnforcementActionRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)

	// Service
	svc := service.NewModerationService(itemRepo, reportRepo, enforcementRepo, auditRepo)

	h := handler.New(svc)

	auth := middleware.JWTAuth(jm)
	modAdmin := middleware.RequireRole("moderator", "admin")
	adminOnly := middleware.RequireRole("admin")

	group := api.Group("/moderation")

	// Moderation queue (auth + moderator/admin)
	group.Get("/queue", auth, modAdmin, h.ListQueue)
	group.Get("/items/:id", auth, modAdmin, h.GetItem)
	group.Put("/items/:id/approve", auth, modAdmin, h.ApproveItem)
	group.Put("/items/:id/reject", auth, modAdmin, h.RejectItem)

	// Reports
	group.Post("/reports", auth, h.CreateReport)
	group.Get("/reports", auth, modAdmin, h.ListReports)
	group.Put("/reports/:id/status", auth, modAdmin, h.UpdateReportStatus)

	// Enforcement actions (create = admin only, list = moderator/admin)
	group.Post("/enforcement", auth, adminOnly, h.CreateEnforcement)
	group.Get("/enforcement/user/:userId", auth, modAdmin, h.ListEnforcementsByUser)

	// Audit logs (admin only)
	group.Get("/audit", auth, adminOnly, h.ListAuditLogs)
}
