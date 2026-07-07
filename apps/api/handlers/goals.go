package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

type GoalsHandler struct {
	Pool *pgxpool.Pool
}

// GetGoals godoc
// @Summary      Get active goals with progress
// @Tags         goals
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/goals [get]
func (h *GoalsHandler) GetGoals(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	ctx := c.Context()
	goals, err := models.GetActiveGoalsWithProgress(ctx, h.Pool, userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch goals")
	}

	return helpers.Success(c, "Goals retrieved", goals)
}

// CreateGoal godoc
// @Summary      Create a new goal
// @Tags         goals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/goals [post]
func (h *GoalsHandler) CreateGoal(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	var input models.CreateGoalInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if input.Scope == "" || input.Period == "" || input.TargetSeconds <= 0 {
		return helpers.Error(c, fiber.StatusBadRequest, "Missing required fields")
	}

	ctx := c.Context()
	goal, err := models.CreateGoal(ctx, h.Pool, userID, input)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to create goal")
	}

	return helpers.Success(c, "Goal created", goal)
}

// UpdateGoal godoc
// @Summary      Update a goal
// @Tags         goals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/goals/{id} [put]
func (h *GoalsHandler) UpdateGoal(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	goalID := c.Params("id")

	var input models.CreateGoalInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	ctx := c.Context()
	goal, err := models.UpdateGoal(ctx, h.Pool, userID, goalID, input)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to update goal")
	}

	return helpers.Success(c, "Goal updated", goal)
}

// DeleteGoal godoc
// @Summary      Delete a goal
// @Tags         goals
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/goals/{id} [delete]
func (h *GoalsHandler) DeleteGoal(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	goalID := c.Params("id")

	ctx := c.Context()
	if err := models.DeleteGoal(ctx, h.Pool, userID, goalID); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to delete goal")
	}

	return helpers.Success(c, "Goal deleted", nil)
}
