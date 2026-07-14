package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/majoramari/seismic/apps/api/handlers"
	"github.com/majoramari/seismic/apps/api/middleware"
)

func Setup(app *fiber.App, authHandler *handlers.AuthHandler, heartbeatHandler *handlers.HeartbeatHandler, adminHandler *handlers.AdminHandler, statsHandler *handlers.StatsHandler, filtersHandler *handlers.FiltersHandler, importHandler *handlers.ImportHandler, jwtSecret string, pool *pgxpool.Pool) {
	app.Get("/health", handlers.HealthCheck)

	auth := app.Group("/api/auth")
	auth.Get("/verify", authHandler.VerifyMagicLink)
	auth.Post("/complete-signup", authHandler.CompleteSignup)
	auth.Post("/refresh", authHandler.RefreshAccessToken)
	auth.Get("/apikey", middleware.RequireAuth(jwtSecret), authHandler.GetAPIKey)
	auth.Post("/apikey/regenerate", middleware.RequireAuth(jwtSecret), authHandler.RegenerateAPIKey)
	auth.Post("/magic-link", authHandler.RequestMagicLink)
	auth.Get("/me", middleware.RequireAuth(jwtSecret), authHandler.GetMe)
	auth.Get("/check-username", authHandler.CheckUsername)
	auth.Post("/change-email", middleware.RequireAuth(jwtSecret), authHandler.RequestEmailChange)
	auth.Get("/confirm-email-change", authHandler.ConfirmEmailChange)

	app.Post("/api/heartbeat", middleware.HeartbeatRateLimit(), middleware.RequireAPIKey(pool), heartbeatHandler.Receive)

	stats := app.Group("/api/stats", middleware.RequireAuthOrAPIKey(pool, jwtSecret))
	stats.Get("/summary", statsHandler.GetSummary)
	stats.Get("/languages", statsHandler.GetLanguages)
	stats.Get("/heatmap", statsHandler.GetHeatmap)
	stats.Get("/projects", statsHandler.GetProjects)
	stats.Get("/editors", statsHandler.GetEditors)
	stats.Get("/timeline", statsHandler.GetTimeline)
	stats.Get("/os", statsHandler.GetOS)
	stats.Get("/dashboard", statsHandler.GetDashboard)

	projectsHandler := &handlers.ProjectsHandler{Pool: pool}
	app.Post("/api/projects/sync", middleware.RequireAPIKey(pool), projectsHandler.SyncProject)

	projects := app.Group("/api/projects", middleware.RequireAuth(jwtSecret))
	projects.Get("", projectsHandler.GetProjects)
	projects.Get("/", projectsHandler.GetProjects)
	projects.Post("/archive", projectsHandler.SetArchived)

	leaderboardHandler := &handlers.LeaderboardHandler{Pool: pool}
	app.Get("/api/leaderboard", middleware.OptionalAuth(pool, jwtSecret), leaderboardHandler.GetLeaderboard)

	settingsHandler := &handlers.SettingsHandler{Pool: pool}
	app.Get("/api/editor/settings", middleware.RequireAPIKey(pool), settingsHandler.GetEditorSettings)

	settings := app.Group("/api/settings", middleware.RequireAuth(jwtSecret))
	settings.Get("/privacy", settingsHandler.GetPrivacy)
	settings.Get("/badges", settingsHandler.GetBadges)
	settings.Get("/editor", settingsHandler.GetEditorSettings)
	settings.Post("/editor", settingsHandler.UpdateEditorSettings)
	settings.Post("/privacy", settingsHandler.UpdatePrivacy)
	settings.Post("/reset-timers", settingsHandler.ResetTimers)
	settings.Post("/account", settingsHandler.DeleteAccount)

	goalsHandler := &handlers.GoalsHandler{Pool: pool}
	goals := app.Group("/api/goals", middleware.RequireAuth(jwtSecret))
	goals.Get("/", goalsHandler.GetGoals)
	goals.Post("/", goalsHandler.CreateGoal)
	goals.Put("/:id", goalsHandler.UpdateGoal)
	goals.Delete("/:id", goalsHandler.DeleteGoal)

	filters := app.Group("/api/filters", middleware.RequireAuth(jwtSecret))
	filters.Get("/languages", filtersHandler.GetLanguages)
	filters.Get("/projects", filtersHandler.GetProjects)

	profileHandler := &handlers.ProfileHandler{Pool: pool}
	app.Get("/api/users/:username", middleware.OptionalAuth(pool, jwtSecret), profileHandler.GetProfile)
	app.Put("/api/auth/profile", middleware.RequireAuth(jwtSecret), authHandler.UpdateProfile)
	app.Post("/api/settings/badge-visibility", middleware.RequireAuth(jwtSecret), profileHandler.ToggleBadgeVisibility)

	app.Post("/api/admin/grant-badge", middleware.RequireAuth(jwtSecret), adminHandler.GrantBadge)
	app.Post("/api/import/wakatime", middleware.RequireAuth(jwtSecret), importHandler.StartWakaTimeImport)
	app.Post("/api/import/file", middleware.RequireAuth(jwtSecret), importHandler.ImportFromFile)
	app.Get("/api/import/status", middleware.RequireAuth(jwtSecret), importHandler.GetImportStatus)
	// This is a testing route, not for production use
	app.Post("/api/admin/process-sessions", middleware.RequireAuth(jwtSecret), adminHandler.TriggerSessionProcessing)
	app.Post("/api/admin/check-reminders", middleware.RequireAuth(jwtSecret), adminHandler.TriggerGoalReminders)
}
