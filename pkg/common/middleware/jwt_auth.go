package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"youtube-code-backend/pkg/common/errors"
	"youtube-code-backend/pkg/common/jwt"
	"youtube-code-backend/pkg/common/response"
)

// JWTAuth validates the Authorization Bearer token.
func JWTAuth(jm *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return response.Err(c, errors.ErrUnauthorized)
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return response.Err(c, errors.ErrTokenInvalid)
		}

		claims, err := jm.ValidateToken(parts[1])
		if err != nil {
			return response.Err(c, errors.ErrTokenInvalid)
		}

		c.Locals("userID", claims.UserID)
		c.Locals("userRole", claims.Role)
		c.Locals("claims", claims)
		return c.Next()
	}
}

// OptionalJWTAuth tries to parse the token but doesn't require it.
func OptionalJWTAuth(jm *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Next()
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			return c.Next()
		}
		claims, err := jm.ValidateToken(parts[1])
		if err != nil {
			return c.Next()
		}
		c.Locals("userID", claims.UserID)
		c.Locals("userRole", claims.Role)
		c.Locals("claims", claims)
		return c.Next()
	}
}

// RequireRole checks if the authenticated user has the required role.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("userRole").(string)
		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}
		return response.Err(c, errors.ErrForbidden)
	}
}

// GetUserID extracts the user ID from context.
func GetUserID(c *fiber.Ctx) uint64 {
	id, _ := c.Locals("userID").(uint64)
	return id
}
