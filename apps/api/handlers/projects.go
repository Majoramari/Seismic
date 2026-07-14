package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

type ProjectsHandler struct {
	Pool *pgxpool.Pool
}

type ProjectArchiveRequest struct {
	Project  string `json:"project"`
	Archived bool   `json:"archived"`
}

func (h *ProjectsHandler) GetProjects(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	rangeParam := c.Query("range", "all")
	archived := c.Query("archived", "false") == "true"

	projects, err := models.GetProjectOverviews(c.Context(), h.Pool, userID, archived, models.RangeSQL(rangeParam))
	if err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to fetch projects")
	}

	return helpers.Success(c, "Projects retrieved", projects)
}

func (h *ProjectsHandler) SyncProject(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var req models.ProjectSync
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if req.Project == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "Project is required")
	}
	if len(req.Commits) > 25 {
		req.Commits = req.Commits[:25]
	}

	if err := models.SyncProject(c.Context(), h.Pool, userID, req); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to sync project")
	}

	return helpers.Success(c, "Project synced", nil)
}

func (h *ProjectsHandler) SetArchived(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var req ProjectArchiveRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if req.Project == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "Project is required")
	}

	if err := models.SetProjectArchived(c.Context(), h.Pool, userID, req.Project, req.Archived); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to update project")
	}

	return helpers.Success(c, "Project updated", nil)
}
