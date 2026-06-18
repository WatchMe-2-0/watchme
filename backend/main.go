package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"watchme/config"
	"watchme/middleware"
	"watchme/models"
	"watchme/utils"

	"github.com/gofiber/fiber/v2"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🎬 WATCHME — Starting up...")

	// ── Load Configuration ──────────────────────────────────────────
	configDir, _ := os.UserHomeDir()
	configPath := filepath.Join(configDir, "watchme", "config.json")
	cfg := config.LoadConfig(configPath)

	// Generate JWT secret if not set
	if cfg.JWTSecret == "" {
		secret, err := utils.GenerateJWTSecret()
		if err != nil {
			log.Fatal("❌ Failed to generate JWT secret:", err)
		}
		config.UpdateConfig(func(c *config.AppConfig) {
			c.JWTSecret = secret
		})
		log.Println("🔑 Generated new JWT secret")
	}

	// Ensure storage directories exist
	if err := config.EnsureDirectories(); err != nil {
		log.Fatal("❌ Failed to create storage directories:", err)
	}

	// ── Database ────────────────────────────────────────────────────
	config.ConnectDatabase(cfg)
	config.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.Movie{},
		&models.Download{},
		&models.Session{},
	)

	// ── Fiber App ───────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		BodyLimit:             5 * 1024 * 1024 * 1024, // 5GB upload limit
		StreamRequestBody:     true,                    // Stream large uploads
		DisableStartupMessage: false,
		AppName:               "WATCHME",
	})

	// ── Middleware ──────────────────────────────────────────────────
	app.Use(middleware.SetupCORS())
	app.Use(middleware.Logger())

	// ── API Routes ─────────────────────────────────────────────────
	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return utils.Success(c, "WATCHME is running", fiber.Map{
			"version": "2.0.0",
			"status":  "healthy",
		})
	})

	// Auth status — check if admin exists (no auth required)
	api.Get("/auth/status", func(c *fiber.Ctx) error {
		var count int64
		config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
		return utils.SuccessData(c, fiber.Map{
			"setup_complete": count > 0,
		})
	})

	// Placeholder route groups (will be implemented in Part 2+)
	// api.Post("/auth/setup", ...)
	// api.Post("/auth/login", ...)
	// api.Get("/profiles", ...)
	// api.Get("/movies", ...)
	// api.Get("/stream/:id", ...)
	// api.Post("/downloads", ...)
	// api.Get("/tmdb/trending", ...)

	// ── Start Server ───────────────────────────────────────────────
	port := cfg.ServerPort
	fmt.Printf("\n")
	fmt.Printf("  ╔══════════════════════════════════════╗\n")
	fmt.Printf("  ║        🎬 WATCHME v2.0.0             ║\n")
	fmt.Printf("  ║   http://localhost:%-18s ║\n", port)
	fmt.Printf("  ╚══════════════════════════════════════╝\n")
	fmt.Printf("\n")

	log.Fatal(app.Listen(":" + port))
}
