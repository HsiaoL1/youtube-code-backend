package response

import (
	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/types"

	"github.com/gofiber/fiber/v2"
)

// Envelope is the standard JSON response wrapper.
type Envelope struct {
	Success    bool                  `json:"success"`
	Data       any                   `json:"data,omitempty"`
	Error      *errors.AppError      `json:"error,omitempty"`
	Pagination *types.PaginationMeta `json:"pagination,omitempty"`
}

// OK sends a 200 JSON response.
func OK(c *fiber.Ctx, data any) error {
	return c.JSON(Envelope{Success: true, Data: data})
}

// Created sends a 201 JSON response.
func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Success: true, Data: data})
}

// NoContent sends a 204 response.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Paginated sends a 200 JSON response with pagination metadata.
func Paginated(c *fiber.Ctx, data any, meta types.PaginationMeta) error {
	return c.JSON(Envelope{Success: true, Data: data, Pagination: &meta})
}

// Err sends an error response based on AppError.
func Err(c *fiber.Ctx, err *errors.AppError) error {
	return c.Status(err.StatusCode).JSON(Envelope{Success: false, Error: err})
}
