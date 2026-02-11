package server

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/config"
	"youtube-code-backend/pkg/common/database"
	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/redis"
	"youtube-code-backend/pkg/common/response"
)

// Resources holds shared infrastructure available to all services.
type Resources struct {
	DB         *gorm.DB
	Redis      *goredis.Client
	JWTManager *jwt.Manager
	Config     config.AppConfig
}

// RegisterRoutes wires service-specific HTTP routes.
type RegisterRoutes func(api fiber.Router, res *Resources)

// Run starts a Fiber app with shared middleware, infra connections, and health endpoints.
func Run(cfg config.AppConfig, register RegisterRoutes) error {
	// Connect to PostgreSQL
	db, err := database.Connect(cfg.DatabaseDSN)
	if err != nil {
		log.Printf("warning: database connection failed: %v", err)
	}

	// Connect to Redis
	rdb, err := redis.Connect(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("warning: redis connection failed: %v", err)
	}

	// Create JWT manager
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessTokenExpiry, cfg.JWTRefreshTokenExpiry)

	res := &Resources{
		DB:         db,
		Redis:      rdb,
		JWTManager: jwtManager,
		Config:     cfg,
	}

	app := fiber.New(fiber.Config{
		AppName:      cfg.ServiceName,
		ErrorHandler: globalErrorHandler,
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
		status := "ready"
		checks := fiber.Map{}

		if db != nil {
			if err := database.Ping(db); err != nil {
				status = "not_ready"
				checks["database"] = "down"
			} else {
				checks["database"] = "up"
			}
		}

		if rdb != nil {
			if err := redis.Ping(rdb); err != nil {
				status = "not_ready"
				checks["redis"] = "down"
			} else {
				checks["redis"] = "up"
			}
		}

		code := fiber.StatusOK
		if status != "ready" {
			code = fiber.StatusServiceUnavailable
		}

		return c.Status(code).JSON(fiber.Map{
			"service": cfg.ServiceName,
			"status":  status,
			"checks":  checks,
		})
	})

	api := app.Group("/api/v1")
	register(api, res)

	return app.Listen(fmt.Sprintf(":%d", cfg.Port))
}

func globalErrorHandler(c *fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		return c.Status(e.Code).JSON(response.Envelope{
			Success: false,
			Error:   errors.New("HTTP_ERR", e.Message, e.Code),
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(response.Envelope{
		Success: false,
		Error:   errors.ErrInternal,
	})
}
