package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"youtube-code-backend/pkg/common/config"
)

// RegisterRoutes wires service-specific HTTP routes.
type RegisterRoutes func(api fiber.Router)

// Run starts a Fiber app with shared middleware and health endpoints.
func Run(cfg config.AppConfig, register RegisterRoutes) error {
	app := fiber.New(fiber.Config{
		AppName: cfg.ServiceName,
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": cfg.ServiceName,
			"status":  "ok",
		})
	})

	app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": cfg.ServiceName,
			"status":  "ready",
		})
	})

	api := app.Group("/api/v1")
	register(api)

	return app.Listen(fmt.Sprintf(":%d", cfg.Port))
}
