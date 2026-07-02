package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/majoramari/seismic/apps/api/config"
	"github.com/majoramari/seismic/apps/api/db"
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

	log.Printf("Seismic API starting on port %s\n", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
