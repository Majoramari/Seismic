package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

type SettingsHandler struct {
	Pool *pgxpool.Pool
}

// GetPrivacy godoc
// @Summary      Get privacy settings
// @Tags         settings
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/settings/privacy [get]
func (h *SettingsHandler) GetPrivacy(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	settings, err := models.GetPrivacySettings(ctx, h.Pool, userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to load settings")
	}

	return helpers.Success(c, "Privacy settings retrieved", settings)
}

func (h *SettingsHandler) GetBadges(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	badges, err := models.GetUserBadgeVisibility(ctx, h.Pool, userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to load badges")
	}

	return helpers.Success(c, "Badges retrieved", badges)
}

// UpdatePrivacy godoc
// @Summary      Update privacy settings
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/settings/privacy [post]
func (h *SettingsHandler) UpdatePrivacy(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var body map[string]bool
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	ctx := c.Context()
	if err := models.UpdatePrivacySettings(ctx, h.Pool, userID, body); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to update settings")
	}

	return helpers.Success(c, "Settings updated", nil)
}

// ResetTimers godoc
// @Summary      Reset all coding stats
// @Tags         settings
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/settings/reset-timers [post]
func (h *SettingsHandler) ResetTimers(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	if err := models.ResetUserTimers(ctx, h.Pool, userID); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to reset timers")
	}

	return helpers.Success(c, "Timers reset successfully", nil)
}

// DeleteAccount godoc
// @Summary      Delete account
// @Tags         settings
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/settings/account [post]
func (h *SettingsHandler) DeleteAccount(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	if err := models.DeleteUserAccount(ctx, h.Pool, userID); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to delete account")
	}

	return helpers.Success(c, "Account deleted", nil)
}
