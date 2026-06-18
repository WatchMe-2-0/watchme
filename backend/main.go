package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"watchme/auth"
	"watchme/config"
	"watchme/handlers"
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
		StreamRequestBody:     true,
		DisableStartupMessage: false,
		AppName:               "WATCHME",
	})

	// ── Global Middleware ───────────────────────────────────────────
	app.Use(middleware.SetupCORS())
	app.Use(middleware.Logger())

	// ── API Routes ─────────────────────────────────────────────────
	api := app.Group("/api")

	// Health check (no auth)
	api.Get("/health", func(c *fiber.Ctx) error {
		return utils.Success(c, "WATCHME is running", fiber.Map{
			"version": "2.0.0",
			"status":  "healthy",
		})
	})

	// ── Auth Routes (no auth required) ──────────────────────────────
	authGroup := api.Group("/auth")
	authGroup.Get("/status", auth.HandleAuthStatus)
	authGroup.Post("/setup", auth.HandleSetup)
	authGroup.Post("/login", auth.HandleLogin)
	authGroup.Post("/logout", auth.HandleLogout)
	authGroup.Post("/recovery", auth.HandleRecovery)

	// ── Profile Routes (public for listing, admin for management) ───
	profileGroup := api.Group("/profiles")
	profileGroup.Get("/", auth.HandleListProfiles)         // Public: profile selection screen
	profileGroup.Get("/avatars", auth.HandleGetAvatars)     // Public: available avatars

	// Admin-only profile management
	profileAdmin := profileGroup.Group("/", auth.RequireAuth(), auth.RequireAdmin())
	profileAdmin.Post("/", auth.HandleCreateProfile)
	profileAdmin.Put("/:id", auth.HandleUpdateProfile)
	profileAdmin.Delete("/:id", auth.HandleDeleteProfile)

	// ── Poster Route (no auth — needed for image loading) ───────────
	api.Get("/posters/:name", handlers.HandleGetPoster)

	// ── Authenticated Routes ────────────────────────────────────────
	authenticated := api.Group("/", auth.RequireAuth(), auth.KidsFilter())

	// Password management (admin only)
	authenticated.Put("/auth/password", auth.HandleChangePassword)

	// Movie routes
	authenticated.Get("/movies", handlers.HandleListMovies)
	authenticated.Get("/movies/:id", handlers.HandleGetMovie)
	authenticated.Post("/upload", handlers.HandleUploadMovie)
	authenticated.Delete("/movies/:id", handlers.HandleDeleteMovie)
	authenticated.Get("/stream/:id", handlers.HandleStreamMovie)

	// Settings routes (admin only)
	settingsGroup := authenticated.Group("/settings", auth.RequireAdmin())
	settingsGroup.Get("/", handlers.HandleGetSettings)
	settingsGroup.Put("/", handlers.HandleUpdateSettings)

	// Placeholder routes (will be implemented in Part 4+)
	// authenticated.Post("/downloads", ...)
	// authenticated.Get("/tmdb/trending", ...)

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
