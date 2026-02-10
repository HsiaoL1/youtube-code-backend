package handler

import "github.com/gofiber/fiber/v2"

// Handler contains service-level endpoint handlers.
type Handler struct{}

// New creates a new handler instance.
func New() *Handler {
	return &Handler{}
}

// Ping is a simple liveness endpoint under the service domain.
func (h *Handler) Ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service": "live-service",
		"message": "pong",
	})
}

// TODOExample is a placeholder endpoint for future implementation.
func (h *Handler) TODOExample(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"service": "live-service",
		"message": "implement business logic in internal/handler",
	})
}
