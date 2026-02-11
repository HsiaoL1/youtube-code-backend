package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/response"
)

// RateLimit creates a rate limiting middleware.
func RateLimit(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		LimitReached: func(c *fiber.Ctx) error {
			return response.Err(c, errors.ErrRateLimited)
		},
	})
}
