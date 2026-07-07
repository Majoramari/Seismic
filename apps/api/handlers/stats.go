package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

type StatsHandler struct {
	Pool *pgxpool.Pool
}

// GetSummary godoc
// @Summary      Get stats summary
// @Description  Returns total time, top language, top project, daily average, and current streak for a time range.
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Param        range query string false "today, week, month, or all" default(today)
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/stats/summary [get]
func (h *StatsHandler) GetSummary(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	rangeParam := c.Query("range", "today")

	ctx := c.Context()
	summary, err := models.GetStatsSummary(ctx, h.Pool, userID, models.RangeSQL(rangeParam))
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch stats")
	}

	return helpers.Success(c, "Stats retrieved", summary)
}

// GetLanguages godoc
// @Summary      Get language breakdown
// @Description  Returns time spent per programming language for a time range.
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Param        range query string false "today, week, month, or all" default(week)
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/stats/languages [get]
func (h *StatsHandler) GetLanguages(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	rangeParam := c.Query("range", "week")

	ctx := c.Context()
	stats, err := models.GetLanguageBreakdown(ctx, h.Pool, userID, models.RangeSQL(rangeParam))
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch stats")
	}

	return helpers.Success(c, "Language breakdown retrieved", stats)
}

// GetHeatmap godoc
// @Summary      Get coding heatmap
// @Description  Returns daily coding totals for the last 365 days, like a GitHub contribution graph.
// @Tags         stats
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/stats/heatmap [get]
func (h *StatsHandler) GetHeatmap(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	days, err := models.GetHeatmap(ctx, h.Pool, userID)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch heatmap")
	}

	return helpers.Success(c, "Heatmap retrieved", days)
}

func (h *StatsHandler) GetProjects(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	rangeParam := c.Query("range", "week")

	ctx := c.Context()
	stats, err := models.GetProjectBreakdown(ctx, h.Pool, userID, models.RangeSQL(rangeParam))
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch stats")
	}

	return helpers.Success(c, "Project breakdown retrieved", stats)
}

func (h *StatsHandler) GetEditors(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	rangeParam := c.Query("range", "week")

	ctx := c.Context()
	stats, err := models.GetEditorBreakdown(ctx, h.Pool, userID, models.RangeSQL(rangeParam))
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch stats")
	}

	return helpers.Success(c, "Editor breakdown retrieved", stats)
}

func (h *StatsHandler) GetTimeline(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	ctx := c.Context()
	timeline, err := models.GetTimeline(ctx, h.Pool, userID, 90)
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch timeline")
	}

	return helpers.Success(c, "Timeline retrieved", timeline)
}
