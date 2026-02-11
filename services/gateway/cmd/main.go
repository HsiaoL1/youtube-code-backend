package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"youtube-code-backend/pkg/common/config"
	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/services/gateway/internal/router"
)

func main() {
	cfg, err := config.FromEnv("gateway-service", 8000)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	jm := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessTokenExpiry, cfg.JWTRefreshTokenExpiry)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"message": err.Error()})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	// Mount all routes under /api (no /v1 prefix)
	api := app.Group("/api")
	router.Register(api, db, jm)

	log.Printf("gateway-service starting on :%d", cfg.Port)
	if err := app.Listen(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatalf("service stopped: %v", err)
	}
}
