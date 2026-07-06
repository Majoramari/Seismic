package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/models"
)

// OptionalAuth tries to identify the user from a JWT or API
// key if present, but never blocks the request if there
// isn't one or it's invalid.
func OptionalAuth(pool *pgxpool.Pool, jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return c.Next()
		}

		value := strings.TrimPrefix(header, "Bearer ")
		ctx := c.Context()

		token, err := jwt.Parse(value, func(t *jwt.Token) (any, error) {
			return []byte(jwtSecret), nil
		})
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userID, ok := claims["sub"].(string); ok {
					c.Locals("userID", userID)
					return c.Next()
				}
			}
		}

		user, err := models.FindUserByAPIKey(ctx, pool, value)
		if err == nil && user != nil {
			c.Locals("userID", user.ID)
		}

		return c.Next()
	}
}
