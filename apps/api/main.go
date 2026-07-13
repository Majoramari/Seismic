package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/gofiber/swagger"
	"github.com/majoramari/seismic/apps/api/middleware"

	"github.com/majoramari/seismic/apps/api/config"
	"github.com/majoramari/seismic/apps/api/db"
	_ "github.com/majoramari/seismic/apps/api/docs"
	"github.com/majoramari/seismic/apps/api/handlers"
	"github.com/majoramari/seismic/apps/api/routes"
	"github.com/majoramari/seismic/apps/api/services"
)

// @title Seismic API
// @version 1.0
// @description Backend API for Seismic, a developer time tracking platform.
// @contact.email hello@seismic.icu
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()
	pool := db.Connect(cfg.DatabaseURL)
	defer pool.Close()

	if err := db.RunMigrations(pool); err != nil {
		log.Fatalf("Migration failed: %v\n", err)
	}

	app := fiber.New()

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(middleware.GeneralRateLimit())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	app.Get("/api/docs/*", fiberSwagger.New())

	authHandler := &handlers.AuthHandler{
		Pool:      pool,
		JWTSecret: cfg.JWTSecret,
		EmailCfg: services.EmailConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUser,
			Password: cfg.SMTPPass,
			AppURL:   cfg.AppURL,
		},
		CloudinaryCfg: services.CloudinaryConfig{
			CloudName: cfg.CloudinaryCloudName,
			APIKey:    cfg.CloudinaryAPIKey,
			APISecret: cfg.CloudinaryAPISecret,
		},
	}
	heartbeatHandler := &handlers.HeartbeatHandler{Pool: pool}
	adminHandler := &handlers.AdminHandler{Pool: pool}
	statsHandler := &handlers.StatsHandler{Pool: pool}
	filtersHandler := &handlers.FiltersHandler{Pool: pool}

	routes.Setup(app, authHandler, heartbeatHandler, adminHandler, statsHandler, filtersHandler, cfg.JWTSecret, pool)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := services.ProcessSessions(context.Background(), pool); err != nil {
				log.Println("Session processor error:", err)
			}
		}
	}()

	log.Printf("Seismic API starting on port %s\n", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
