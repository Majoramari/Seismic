package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
	"github.com/majoramari/seismic/apps/api/services"
)

type ImportHandler struct {
	Pool     *pgxpool.Pool
	progress map[string]*importProgress
	mu       sync.Mutex
}

type importProgress struct {
	Status   string `json:"status"` // "processing" | "completed" | "failed"
	Imported int    `json:"imported"`
	Total    int    `json:"total"`
	Error    string `json:"error,omitempty"`
}

type importRequest struct {
	APIKey string `json:"apiKey"`
}

type fileImportRequest struct {
	FileContent string `json:"fileContent"`
}

// StartWakaTimeImport godoc
// @Summary      Import heartbeats from WakaTime
// @Description  Kicks off a background import of the user's full WakaTime heartbeat history. Runs asynchronously; check back later for updated stats.
// @Tags         import
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body importRequest true "WakaTime API key"
// @Success      202 {object} helpers.APIResponse
// @Failure      400 {object} helpers.APIResponse
// @Router       /api/import/wakatime [post]
func (h *ImportHandler) StartWakaTimeImport(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	var body importRequest
	if err := c.BodyParser(&body); err != nil || body.APIKey == "" {
		return helpers.Error(c, fiber.StatusBadRequest, "WakaTime API key is required")
	}

	dumpID, err := services.RequestWakaTimeHeartbeatDump(body.APIKey)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Failed to start WakaTime export. Check your API key.")
	}

	// Everything past this point can take minutes, so it runs in the
	// background — the request returns immediately.
	go func() {
		ctx := context.Background()

		downloadURL, err := services.PollWakaTimeDump(ctx, body.APIKey, dumpID)
		if err != nil {
			log.Printf("WakaTime import failed for user %s: %v", userID, err)
			return
		}

		heartbeats, err := services.DownloadWakaTimeDump(downloadURL)
		if err != nil {
			log.Printf("WakaTime import download/parse failed for user %s: %v", userID, err)
			return
		}

		inserted, err := models.InsertImportedHeartbeats(
			ctx,
			h.Pool,
			userID,
			heartbeats,
			func(count int) {
				h.setProgress(userID, &importProgress{
					Status:   "processing",
					Imported: count,
					Total:    len(heartbeats),
				})
			},
		)
		if err != nil {
			log.Printf("WakaTime import insert failed for user %s: %v", userID, err)
			return
		}

		log.Printf("WakaTime import complete for user %s: %d heartbeats inserted", userID, inserted)

		if err := services.ReprocessUserSessions(ctx, h.Pool, userID); err != nil {
			log.Printf("Session processing after import failed for user %s: %v", userID, err)
		}
	}()

	return helpers.Success(c, "Import started — this may take a few minutes depending on your history size", nil)
}

func (h *ImportHandler) setProgress(userID string, p *importProgress) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.progress == nil {
		h.progress = make(map[string]*importProgress)
	}
	h.progress[userID] = p
}

// GetImportStatus godoc
// @Summary      Get import progress
// @Tags         import
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Router       /api/import/status [get]
func (h *ImportHandler) GetImportStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	h.mu.Lock()
	p, ok := h.progress[userID]
	h.mu.Unlock()

	if !ok {
		return helpers.Success(c, "No import in progress", fiber.Map{"status": "idle"})
	}
	return helpers.Success(c, "Import status", p)
}

// ImportFromFile godoc
// @Summary      Import heartbeats from an uploaded export file
// @Description  Accepts a WakaTime or Hackatime JSON export and imports its heartbeats.
// @Tags         import
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body fileImportRequest true "Raw file content"
// @Success      200 {object} helpers.APIResponse
// @Failure      400 {object} helpers.APIResponse
// @Router       /api/import/file [post]
func (h *ImportHandler) ImportFromFile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "No file provided")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Could not read uploaded file")
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Could not read uploaded file")
	}

	jsonBytes, err := extractJSON(data, fileHeader.Filename)
	if err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, err.Error())
	}

	var parsed models.ImportFile
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Could not parse file — make sure it's a valid WakaTime or Hackatime export")
	}

	heartbeats := models.FlattenImportFile(parsed)
	if len(heartbeats) == 0 {
		return helpers.Error(c, fiber.StatusBadRequest, "No heartbeats found in that file")
	}
	if parsed.ExportInfo.TotalHeartbeats > 0 && len(heartbeats) < parsed.ExportInfo.TotalHeartbeats {
		return helpers.Error(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf(
				"This export only contains %d of %d heartbeats. Download the full Hackatime export, not a paginated or preview file.",
				len(heartbeats),
				parsed.ExportInfo.TotalHeartbeats,
			),
		)
	}

	h.setProgress(userID, &importProgress{Status: "processing", Total: len(heartbeats)})

	go func() {
		ctx := context.Background()

		inserted, err := models.InsertImportedHeartbeats(ctx, h.Pool, userID, heartbeats, func(count int) {
			h.setProgress(userID, &importProgress{Status: "processing", Imported: count, Total: len(heartbeats)})
		})
		if err != nil {
			h.setProgress(userID, &importProgress{Status: "failed", Error: err.Error()})
			return
		}

		h.setProgress(userID, &importProgress{Status: "completed", Imported: inserted, Total: len(heartbeats)})

		if err := services.ReprocessUserSessions(ctx, h.Pool, userID); err != nil {
			log.Printf("Session processing after file import failed for user %s: %v", userID, err)
		}
	}()

	return helpers.Success(c, "Import started", fiber.Map{"total": len(heartbeats)})
}

// extractJSON returns the raw JSON to parse, unzipping first if the
// uploaded file is a .zip archive (Hackatime exports as a zip
// containing a single JSON file).
func extractJSON(data []byte, filename string) ([]byte, error) {
	if !strings.HasSuffix(strings.ToLower(filename), ".zip") {
		return data, nil
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("could not open zip file")
	}

	for _, f := range reader.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".json") {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("could not read %s from the zip", f.Name)
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}

	return nil, fmt.Errorf("no JSON file found inside the zip archive")
}
