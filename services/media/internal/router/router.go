package router

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/services/media/internal/handler"
)

// Register wires service routes under /api/v1.
func Register(api fiber.Router) {
	h := handler.New()

	group := api.Group("/media")
	group.Get("/ping", h.Ping)
	group.Get("/todo", h.TODOExample)
}
