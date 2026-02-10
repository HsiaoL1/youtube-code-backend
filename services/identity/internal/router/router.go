package router

import (
	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/services/identity/internal/handler"
)

// Register wires service routes under /api/v1.
func Register(api fiber.Router) {
	h := handler.New()

	group := api.Group("/identity")
	group.Get("/ping", h.Ping)
	group.Get("/todo", h.TODOExample)
}
