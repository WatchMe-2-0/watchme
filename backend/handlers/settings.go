package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"watchme/config"
	"watchme/utils"
)

// SettingsResponse is the public settings view
type SettingsResponse struct {
	DownloadDir               string `json:"download_dir"`
	PosterDir                 string `json:"poster_dir"`
	TMDBApiKey                string `json:"tmdb_api_key"`
	MaxConcurrentDownloads    int    `json:"max_concurrent_downloads"`
	SessionExpiry             int    `json:"session_expiry_days"`
	EnableStreamWhileDownload bool   `json:"enable_stream_while_download"`
	ServerPort                string `json:"server_port"`
}

// HandleGetSettings returns current app settings (admin only)
func HandleGetSettings(c *fiber.Ctx) error {
	cfg := config.Get()
	return utils.SuccessData(c, SettingsResponse{
		DownloadDir:               cfg.DownloadDir,
		PosterDir:                 cfg.PosterDir,
		TMDBApiKey:                maskAPIKey(cfg.TMDBApiKey),
		MaxConcurrentDownloads:    cfg.MaxConcurrentDownloads,
		SessionExpiry:             cfg.SessionExpiry,
		EnableStreamWhileDownload: cfg.EnableStreamWhileDownload,
		ServerPort:                cfg.ServerPort,
	})
}

// HandleUpdateSettings updates app settings (admin only)
func HandleUpdateSettings(c *fiber.Ctx) error {
	var req struct {
		DownloadDir               *string `json:"download_dir,omitempty"`
		PosterDir                 *string `json:"poster_dir,omitempty"`
		TMDBApiKey                *string `json:"tmdb_api_key,omitempty"`
		MaxConcurrentDownloads    *int    `json:"max_concurrent_downloads,omitempty"`
		SessionExpiry             *int    `json:"session_expiry_days,omitempty"`
		EnableStreamWhileDownload *bool   `json:"enable_stream_while_download,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	err := config.UpdateConfig(func(cfg *config.AppConfig) {
		if req.DownloadDir != nil {
			cfg.DownloadDir = *req.DownloadDir
		}
		if req.PosterDir != nil {
			cfg.PosterDir = *req.PosterDir
		}
		if req.TMDBApiKey != nil {
			cfg.TMDBApiKey = *req.TMDBApiKey
		}
		if req.MaxConcurrentDownloads != nil && *req.MaxConcurrentDownloads > 0 {
			cfg.MaxConcurrentDownloads = *req.MaxConcurrentDownloads
		}
		if req.SessionExpiry != nil && *req.SessionExpiry > 0 {
			cfg.SessionExpiry = *req.SessionExpiry
		}
		if req.EnableStreamWhileDownload != nil {
			cfg.EnableStreamWhileDownload = *req.EnableStreamWhileDownload
		}
	})

	if err != nil {
		log.Printf("❌ Failed to update settings: %v", err)
		return utils.InternalError(c, "Failed to save settings")
	}

	// Re-ensure directories after potential change
	if err := config.EnsureDirectories(); err != nil {
		log.Printf("⚠️  Failed to create new directories: %v", err)
	}

	log.Println("✅ Settings updated")
	return utils.Success(c, "Settings updated successfully", nil)
}

// maskAPIKey shows only the last 4 characters of an API key
func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}
