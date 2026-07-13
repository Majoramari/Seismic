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

	// viewerID is empty for anonymous visitors — set only when the
	// optional auth middleware successfully validated a token.
	viewerID, _ := c.Locals("userID").(string)

	ctx := c.Context()
	profile, err := models.GetPublicProfile(ctx, h.Pool, username, viewerID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
	if profile == nil {
		return helpers.Error(c, fiber.StatusNotFound, "User not found")
	}

	return helpers.Success(c, "Profile retrieved", profile)
}
