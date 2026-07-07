package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

type FiltersHandler struct {
	Pool *pgxpool.Pool
}

// GetLanguages godoc
// @Summary      Get distinct languages coded
// @Tags         filters
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/filters/languages [get]
func (h *FiltersHandler) GetLanguages(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	ctx := c.Context()

	languages, err := models.GetDistinctLanguages(ctx, h.Pool, userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch languages")
	}
	return helpers.Success(c, "Languages retrieved", languages)
}

// GetProjects godoc
// @Summary      Get distinct projects coded
// @Tags         filters
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/filters/projects [get]
func (h *FiltersHandler) GetProjects(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	ctx := c.Context()

	projects, err := models.GetDistinctProjects(ctx, h.Pool, userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch projects")
	}
	return helpers.Success(c, "Projects retrieved", projects)
}
