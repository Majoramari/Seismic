package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

type ProfileHandler struct {
	Pool *pgxpool.Pool
}

// GetProfile godoc
// @Summary      Get public profile
// @Description  Returns a user's public profile, respecting their privacy settings.
// @Tags         profile
// @Produce      json
// @Param        username path string true "Username"
// @Success      200 {object} helpers.APIResponse
// @Failure      404 {object} helpers.APIResponse
// @Router       /api/users/{username} [get]
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	username := c.Params("username")

	ctx := c.Context()
	profile, err := models.GetPublicProfile(ctx, h.Pool, username)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
	if profile == nil {
		return helpers.Error(c, fiber.StatusNotFound, "User not found")
	}

	return helpers.Success(c, "Profile retrieved", profile)
}
