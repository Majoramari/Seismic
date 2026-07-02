package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/majoramari/seismic/apps/api/config"
	"github.com/majoramari/seismic/apps/api/db"
	"github.com/majoramari/seismic/apps/api/handlers"
	"github.com/majoramari/seismic/apps/api/services"
)

func main() {
	cfg := config.Load()
	pool := db.Connect(cfg.DatabaseURL)
	defer pool.Close()

	if err := db.RunMigrations(pool); err != nil {
		log.Fatalf("Migration failed: %v\n", err)
	}

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	authHandler := &handlers.AuthHandler{
		Pool: pool,
		EmailCfg: services.EmailConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUser,
			Password: cfg.SMTPPass,
			AppURL:   cfg.AppURL,
		},
	}

	app.Post("/api/auth/magic-link", authHandler.RequestMagicLink)

	log.Printf("Seismic API starting on port %s\n", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
