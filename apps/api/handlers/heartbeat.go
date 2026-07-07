package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

type HeartbeatHandler struct {
	Pool *pgxpool.Pool
}

// Receive godoc
// @Summary      Send a heartbeat
// @Description  Called by editor plugins every 2 minutes to record coding activity. Requires an API key, not a JWT.
// @Tags         heartbeat
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.Heartbeat true "Heartbeat payload"
// @Success      200 {object} helpers.APIResponse
// @Failure      400 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/heartbeat [post]
func (h *HeartbeatHandler) Receive(c *fiber.Ctx) error {
	var hb models.Heartbeat
	if err := c.BodyParser(&hb); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if hb.File == "" || hb.Project == "" || hb.Language == "" || hb.Editor == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "Missing required fields")
	}

	now := time.Now().UnixMilli()
	if hb.Time > now+2*60*1000 {
		return helpers.Error(c, fiber.StatusBadRequest, "Timestamp too far in the future")
	}
	if hb.Time < now-5*60*1000 {
		return helpers.Error(c, fiber.StatusBadRequest, "Timestamp too far in the past")
	}

	userID, ok := c.Locals("userID").(string)
	if !ok {
		return helpers.Error(c, fiber.StatusUnauthorized, "Unauthorized")
	}
	ctx := c.Context()

	isDuplicate, err := models.HasRecentDuplicate(ctx, h.Pool, userID, hb.File, hb.Time)
	if err != nil {
		log.Println("HasRecentDuplicate error:", err)
		return helpers.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
	if isDuplicate {
		return helpers.Success(c, "Duplicate heartbeat ignored", nil)
	}

	if err := models.InsertHeartbeat(ctx, h.Pool, userID, hb); err != nil {
		log.Println("InsertHeartbeat error:", err)
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to save heartbeat")
	}

	return helpers.Success(c, "Heartbeat received", nil)
}
