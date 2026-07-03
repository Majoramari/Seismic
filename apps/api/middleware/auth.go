package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/majoramari/seismic/apps/api/helpers"
)

// RequireAuth checks for a valid access token in the
// Authorization header and attaches the user ID to the
// request context as "userID" for handlers to use.
func RequireAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return helpers.Error(c, fiber.StatusUnauthorized, "Missing or invalid authorization header")
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return helpers.Error(c, fiber.StatusUnauthorized, "Invalid or expired access token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return helpers.Error(c, fiber.StatusUnauthorized, "Invalid token claims")
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			return helpers.Error(c, fiber.StatusUnauthorized, "Invalid token subject")
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}
