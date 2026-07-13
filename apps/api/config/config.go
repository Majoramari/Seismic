package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all the settings that seismic needs, loaded at startup from the env vars
type Config struct {
	DatabaseURL         string
	JWTSecret           string
	Port                string
	SMTPHost            string
	SMTPPort            string
	SMTPUser            string
	SMTPPass            string
	AppURL              string
	AllowedOrigins      string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

// Load reads the .env file (if exist) and returns a Config struct
// with all values filled in. It exits immediately if a required value is missing
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment directly")
	}

	cfg := &Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		Port:                os.Getenv("PORT"),
		SMTPHost:            os.Getenv("SMTP_HOST"),
		SMTPPort:            os.Getenv("SMTP_PORT"),
		SMTPUser:            os.Getenv("SMTP_USER"),
		SMTPPass:            os.Getenv("SMTP_PASS"),
		AppURL:              os.Getenv("APP_URL"),
		AllowedOrigins:      os.Getenv("ALLOWED_ORIGINS"),
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: os.Getenv("CLOUDINARY_API_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set in .env")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is not set in .env")
	}
	if cfg.Port == "" {
		cfg.Port = "5024"
	}
	if cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPass == "" {
		log.Println("Warning: SMTP settings are incomplete, magic link emails will fail")
	}

	return cfg
}
