package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/majoramari/seismic/apps/api/helpers"
)

// HeartbeatRateLimit allows normal editor activity, quick file
// switches, and queued retry flushes. Duplicate heartbeat checks
// handle noisy clients; this is only a broad abuse guard.
func HeartbeatRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("Authorization") // rate limit per api key, not per IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return helpers.Error(c, fiber.StatusTooManyRequests, "Too many heartbeats, slow down")
		},
	})
}

// AuthRateLimit allows at most 3 requests per hour per IP,
// used on the magic link endpoint to prevent email spam abuse.
func AuthRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,
		Expiration: time.Hour,
		LimitReached: func(c *fiber.Ctx) error {
			return helpers.Error(c, fiber.StatusTooManyRequests, "Too many requests, try again later")
		},
	})
}

// GeneralRateLimit allows at most 100 requests per minute per
// IP, a broad safety net for all other endpoints.
func GeneralRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return helpers.Error(c, fiber.StatusTooManyRequests, "Too many requests")
		},
	})
}
