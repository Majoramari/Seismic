package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

// RequireAPIKey checks for a valid API key in the
// Authorization header and attaches the user ID to context.
// Used by our plugins, unlike RequireAuth which uses JWTs.
func RequireAPIKey(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return helpers.Error(c, fiber.StatusUnauthorized, "Missing api key")
		}

		apiKey := strings.TrimPrefix(header, "Bearer ")

		ctx := c.Context()
		user, err := models.FindUserByAPIKey(ctx, pool, apiKey)
		if err != nil {
			log.Println("FindUserByAPIKey error:", err)
			return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
		}
		if user == nil {
			return helpers.Error(c, fiber.StatusUnauthorized, "Invalid api key")
		}

		c.Locals("userID", user.ID)
		return c.Next()
	}
}
