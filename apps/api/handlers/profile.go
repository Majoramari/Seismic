package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/helpers"
	"github.com/majoramari/seismic/apps/api/models"
)

// ProfileHandler handles requests related to the user's profile page.
type ProfileHandler struct {
	Pool *pgxpool.Pool
}

// GetProfile godoc
// @Summary      Get full profile
// @Description  Returns the logged-in user's full profile: bio info, coding metrics, activity heatmap, problem-solving gauge, achievements, recent activity, and personal-info completion.
// @Tags         profile
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Failure      401 {object} helpers.APIResponse
// @Router       /api/profile [get]
func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	ctx := c.Context()



	// ── 1. User row ──────────────────────────────────────────────────────────
	user, err := models.FindUserByID(ctx, h.Pool, userID)
	if err != nil || user == nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to load user")
	}

	// ── 2. Coding stats (cached aggregate) ──────────────────────────────────
	stats, _ := models.GetProfileStats(ctx, h.Pool, userID)
	if stats == nil {
		stats = &models.UserProfileStats{}
	}

	// ── 3. Problem-solving stats (placeholder — zeros if no row) ────────────
	problemStats, _ := models.GetProblemStats(ctx, h.Pool, userID)
	if problemStats == nil {
		problemStats = &models.UserProblemStats{}
	}

	// ── 4. Achievements ──────────────────────────────────────────────────────
	achievements, _ := models.GetProfileAchievements(ctx, h.Pool, userID)
	if achievements == nil {
		achievements = []models.Achievement{}
	}

	// ── 5. Recent activity ───────────────────────────────────────────────────
	activity, _ := models.GetRecentActivity(ctx, h.Pool, userID, 10)
	if activity == nil {
		activity = []models.ActivityLogItem{}
	}

	// ── 6. Heatmap: flat days → 52-week × 7-day grid ────────────────────────
	heatmapDays, _ := models.GetHeatmap(ctx, h.Pool, userID)
	heatmap := buildHeatmapGrid(heatmapDays)

	// ── 7. Personal-info completion ──────────────────────────────────────────
	completionPercent, infoFields := computeInfoCompletion(user)

	// ── 8. Derived string fields ─────────────────────────────────────────────
	firstName := derefStr(user.FirstName)
	role := derefStr(user.Role)
	location := derefStr(user.Location)
	university := derefStr(user.University)
	bio := derefStr(user.Bio)
	timeZone := derefStr(user.TimeZone)
	website := derefStr(user.Website)
	gender := derefStr(user.Gender)
	languages := user.Languages
	if languages == nil {
		languages = []string{}
	}
	avatarURL := derefStr(user.AvatarURL)

	joinDate := user.CreatedAt.Format("Jan 2006")
	memberFor := computeMemberFor(user.CreatedAt)
	lastActive := computeLastActive(user.LastActiveAt)

	resp := models.ProfileResponse{
		Username:  user.Username,
		Email:     user.Email,
		FirstName: firstName,
		AvatarURL: avatarURL,

		Role:       role,
		Location:   location,
		University: university,
		Bio:        bio,
		TimeZone:   timeZone,
		Website:    website,
		Gender:     gender,
		Languages:  languages,

		JoinDate:   joinDate,
		LastActive: lastActive,
		MemberFor:  memberFor,

		TotalCodingSeconds: stats.TotalCodingSeconds,
		TotalActiveDays:    stats.TotalActiveDays,
		CurrentStreak:      stats.CurrentStreak,
		MaxStreak:          stats.MaxStreak,

		// Placeholder — see migration 010 comment
		Solved:        problemStats.SolvedCount,
		TotalProblems: problemStats.TotalProblems,
		Attempting:    problemStats.AttemptingCount,

		CompletionPercent: completionPercent,
		InfoFields:        infoFields,

		Heatmap:        heatmap,
		RecentActivity: activity,
		Achievements:   achievements,
	}

	return helpers.Success(c, "Profile retrieved", resp)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// derefStr safely dereferences a *string; returns "" if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// computeMemberFor returns a human-readable duration since createdAt,
// e.g. "2 years", "8 months", "3 weeks".
func computeMemberFor(createdAt time.Time) string {
	now := time.Now()
	years := now.Year() - createdAt.Year()
	months := int(now.Month()) - int(createdAt.Month())
	if now.Day() < createdAt.Day() {
		months--
	}
	if months < 0 {
		years--
		months += 12
	}
	totalMonths := years*12 + months
	switch {
	case totalMonths >= 12:
		y := totalMonths / 12
		if y == 1 {
			return "1 year"
		}
		return fmt.Sprintf("%d years", y)
	case totalMonths > 0:
		if totalMonths == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", totalMonths)
	default:
		days := int(now.Sub(createdAt).Hours() / 24)
		if days < 7 {
			return "just joined"
		}
		return fmt.Sprintf("%d weeks", days/7)
	}
}

// computeLastActive returns a human-readable string for when the user was
// last active. Returns "Online now" if within the last 5 minutes.
func computeLastActive(lastActive *time.Time) string {
	if lastActive == nil {
		return "Never"
	}
	diff := time.Since(*lastActive)
	switch {
	case diff < 5*time.Minute:
		return "Online now"
	case diff < time.Hour:
		m := int(diff.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case diff < 24*time.Hour:
		h := int(diff.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		d := int(diff.Hours() / 24)
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	}
}

// computeInfoCompletion checks which profile fields are filled and returns
// the completion percentage and the list of info fields for the UI capsules.
// Fields checked: fullName (display_name), bio, location, gender, languages.
func computeInfoCompletion(user *models.User) (int, []models.ProfileInfoField) {
	type check struct {
		key   string
		label string
		ok    bool
	}

	// display_name used as "Full name" since there's no separate fullName column
	displayNameFilled := user.DisplayName != nil && *user.DisplayName != ""
	bioFilled := user.Bio != nil && *user.Bio != ""
	locationFilled := user.Location != nil && *user.Location != ""
	genderFilled := user.Gender != nil && *user.Gender != ""
	languagesFilled := len(user.Languages) > 0

	checks := []check{
		{"fullName", "Full name", displayNameFilled},
		{"bio", "Bio", bioFilled},
		{"location", "Location", locationFilled},
		{"gender", "Gender", genderFilled},
		{"languages", "Languages", languagesFilled},
	}

	completed := 0
	fields := make([]models.ProfileInfoField, 0, len(checks))
	for _, c := range checks {
		if c.ok {
			completed++
		}
		fields = append(fields, models.ProfileInfoField{
			Key:       c.key,
			Label:     c.label,
			Completed: c.ok,
		})
	}

	pct := (completed * 100) / len(checks)
	return pct, fields
}

// buildHeatmapGrid converts a flat []HeatmapDay (from GetHeatmap) into a
// 52-week × 7-day grid suitable for the frontend contribution chart.
// Level thresholds (seconds): 0→0, 1-1799→1, 1800-3599→2, 3600-7199→3, ≥7200→4.
func buildHeatmapGrid(days []models.HeatmapDay) [][]models.HeatmapCell {
	// Build a lookup: date-string → seconds
	lookup := make(map[string]int, len(days))
	for _, d := range days {
		lookup[d.Date] = d.Seconds
	}

	// Anchor: today, walk back to find the start of the grid (52 weeks ago).
	// We start on the same weekday as 364 days ago (52 full weeks).
	today := time.Now().Truncate(24 * time.Hour)
	startDay := today.AddDate(0, 0, -364) // 52*7 = 364 days

	grid := make([][]models.HeatmapCell, 52)
	for w := 0; w < 52; w++ {
		week := make([]models.HeatmapCell, 7)
		for d := 0; d < 7; d++ {
			t := startDay.AddDate(0, 0, w*7+d)
			dateStr := t.Format("2006-01-02")
			sec := lookup[dateStr]
			week[d] = models.HeatmapCell{
				Date:  dateStr,
				Level: secondsToLevel(sec),
			}
		}
		grid[w] = week
	}
	return grid
}

// secondsToLevel maps coding seconds for a day to a display level 0-4.
func secondsToLevel(sec int) int {
	switch {
	case sec <= 0:
		return 0
	case sec < 1800: // < 30 min
		return 1
	case sec < 3600: // 30 min – 1 h
		return 2
	case sec < 7200: // 1 h – 2 h
		return 3
	default: // ≥ 2 h
		return 4
	}
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates allowed fields on the user profile
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} helpers.APIResponse
// @Failure      400 {object} helpers.APIResponse
// @Router       /api/profile [patch]
func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	ctx := c.Context()

	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.Error(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validation: Bio length
	if req.Bio != nil && len(*req.Bio) > 280 {
		return helpers.Error(c, fiber.StatusBadRequest, "Bio must not exceed 280 characters")
	}

	// Validation: Website URL format (only if provided and not empty)
	if req.Website != nil && *req.Website != "" {
		if _, err := url.ParseRequestURI(*req.Website); err != nil {
			return helpers.Error(c, fiber.StatusBadRequest, "Invalid website URL format")
		}
	}

	if err := models.UpdateUserProfile(ctx, h.Pool, userID, &req); err != nil {
		return helpers.Error(c, fiber.StatusInternalServerError, "Failed to update profile")
	}

	return helpers.Success(c, "Profile updated successfully", nil)
}
