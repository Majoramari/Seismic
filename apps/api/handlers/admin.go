package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/majoramari/seismic/apps/api/models"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/services"
)

type AdminHandler struct {
	Pool     *pgxpool.Pool
	EmailCfg services.EmailConfig
}

type grantBadgeInput struct {
	Username  string `json:"username"`
	BadgeType string `json:"badgeType"`
}

var validRoleBadges = map[string]bool{
	"supporter":   true,
	"contributor": true,
	"maintainer":  true,
}

// TriggerSessionProcessing godoc
// @Summary      Manually trigger session processing
// @Description  Runs the session processor immediately instead of waiting for the 5 minute timer. For testing only.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/admin/process-sessions [post]
func (h *AdminHandler) TriggerSessionProcessing(c *fiber.Ctx) error {
	ctx := c.Context()
	if err := services.ProcessSessions(ctx, h.Pool); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to process sessions")
	}
	return helpers.Success(c, "Sessions processed", nil)
}

func (h *AdminHandler) TriggerGoalReminders(c *fiber.Ctx) error {
	ctx := c.Context()
	if err := services.CheckGoalReminders(ctx, h.Pool, h.EmailCfg); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to check reminders")
	}
	return helpers.Success(c, "Reminders checked", nil)
}

func (h *AdminHandler) GrantBadge(c *fiber.Ctx) error {
	var body grantBadgeInput
	if err := c.BodyParser(&body); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if !validRoleBadges[body.BadgeType] {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid badge type")
	}

	ctx := c.Context()
	user, err := models.FindUserByUsername(ctx, h.Pool, body.Username)
	if err != nil || user == nil {
		return helpers.Error(c, fiber.StatusNotFound, "User not found")
	}

	_, err = h.Pool.Exec(ctx, `
		INSERT INTO badges (user_id, badge_type)
		VALUES ($1, $2)
		ON CONFLICT (user_id, badge_type) DO NOTHING
	`, user.ID, body.BadgeType)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to grant badge")
	}

	return helpers.Success(c, "Badge granted", nil)
}
