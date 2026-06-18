package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"watchme/auth"
	"watchme/config"
	"watchme/handlers"
	"watchme/middleware"
	"watchme/models"
	engine "watchme/torrent"
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

	// ── Torrent Engine ─────────────────────────────────────────────
	torrentEngine := engine.GetEngine()
	if err := torrentEngine.Start(); err != nil {
		log.Printf("⚠️  Torrent engine failed to start: %v", err)
	} else {
		// Start download worker pool
		pool := engine.GetPool()
		pool.Start()

		// Register progress listener to update DB
		torrentEngine.OnProgress(func(update engine.ProgressUpdate) {
			updates := map[string]interface{}{
				"progress":   update.Progress,
				"speed":      update.Speed,
				"peers":      update.Peers,
				"eta":        update.ETA,
				"downloaded":  update.Downloaded,
				"file_size":  update.Total,
				"status":     update.Status,
			}
			if update.FilePath != "" {
				updates["file_path"] = update.FilePath
			}
			config.DB.Table("downloads").Where("id = ?", update.DownloadID).Updates(updates)
		})
	}

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

	// Download routes
	authenticated.Post("/downloads", handlers.HandleStartDownload)
	authenticated.Get("/downloads", handlers.HandleListDownloads)
	authenticated.Delete("/downloads/:id", handlers.HandleCancelDownload)
	authenticated.Get("/downloads/progress", handlers.HandleDownloadProgress)

	// TMDB routes (browse movies)
	tmdbGroup := authenticated.Group("/tmdb")
	tmdbGroup.Get("/search", handlers.HandleTMDBSearch)
	tmdbGroup.Get("/trending", handlers.HandleTMDBTrending)
	tmdbGroup.Get("/top-rated", handlers.HandleTMDBTopRated)
	tmdbGroup.Get("/genre/:id", handlers.HandleTMDBByGenre)
	tmdbGroup.Get("/genres", handlers.HandleTMDBGenres)
	tmdbGroup.Get("/movie/:id", handlers.HandleTMDBMovieDetail)

	// TMDB poster proxy (no auth for image loading)
	api.Get("/tmdb/poster/*", handlers.HandleTMDBPosterProxy)

	// ── Start Server ───────────────────────────────────────────────
	port := cfg.ServerPort
	fmt.Printf("\n")
	fmt.Printf("  ╔══════════════════════════════════════╗\n")
	fmt.Printf("  ║        🎬 WATCHME v2.0.0             ║\n")
	fmt.Printf("  ║   http://localhost:%-18s ║\n", port)
	fmt.Printf("  ╚══════════════════════════════════════╝\n")
	fmt.Printf("\n")

	// ── Graceful Shutdown ──────────────────────────────────────────
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("\n🛑 Shutting down...")
		torrentEngine.Stop()
		app.Shutdown()
	}()

	log.Fatal(app.Listen(":" + port))
}
