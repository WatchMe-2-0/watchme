package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"watchme/auth"
	"watchme/config"
	"watchme/models"
	enginepkg "watchme/torrent"
	"watchme/utils"
)

// HandleStartDownload initiates a torrent download from a magnet link
func HandleStartDownload(c *fiber.Ctx) error {
	var req struct {
		MagnetLink string `json:"magnet_link"`
		Title      string `json:"title"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	if req.MagnetLink == "" {
		return utils.BadRequest(c, "Magnet link is required")
	}

	if !strings.HasPrefix(req.MagnetLink, "magnet:") {
		return utils.BadRequest(c, "Invalid magnet link format")
	}

	profileID := auth.GetProfileID(c)

	// Extract info hash from magnet link
	infoHash := extractInfoHash(req.MagnetLink)
	if infoHash == "" {
		return utils.BadRequest(c, "Could not extract info hash from magnet link")
	}

	// Check if already downloading
	var existing models.Download
	if err := config.DB.Where("info_hash = ? AND status IN ?", infoHash,
		[]string{"queued", "downloading"}).First(&existing).Error; err == nil {
		return utils.BadRequest(c, "This torrent is already being downloaded")
	}

	// Create download record
	download := models.Download{
		ProfileID:  profileID,
		MagnetLink: req.MagnetLink,
		InfoHash:   infoHash,
		Title:      req.Title,
		Status:     models.StatusQueued,
	}

	if err := config.DB.Create(&download).Error; err != nil {
		log.Printf("❌ Failed to create download record: %v", err)
		return utils.InternalError(c, "Failed to start download")
	}

	// Submit to download pool
	pool := enginepkg.GetPool()
	pool.Submit(enginepkg.DownloadRequest{
		MagnetURI:  req.MagnetLink,
		ProfileID:  profileID,
		DownloadID: download.ID,
		Title:      req.Title,
	})

	log.Printf("🧲 Download queued: %s (hash: %s)", req.Title, infoHash[:12])

	return utils.Success(c, "Download started", fiber.Map{
		"id":        download.ID,
		"info_hash": infoHash,
		"title":     req.Title,
		"status":    "queued",
	})
}

// HandleListDownloads returns all downloads for the current profile
func HandleListDownloads(c *fiber.Ctx) error {
	profileID := auth.GetProfileID(c)

	var downloads []models.Download
	query := config.DB.Order("created_at DESC")

	// Admin sees all, regular users see only their own
	claims := auth.GetClaims(c)
	if claims == nil || claims.Role != "admin" {
		query = query.Where("profile_id = ?", profileID)
	}

	query.Find(&downloads)

	// Merge live progress for active downloads
	engine := enginepkg.GetEngine()
	activeDownloads := engine.GetActiveDownloads()
	activeMap := make(map[uint]enginepkg.ProgressUpdate)
	for _, ad := range activeDownloads {
		activeMap[ad.DownloadID] = ad
	}

	type DownloadResponse struct {
		models.Download
		LiveProgress *float64 `json:"live_progress,omitempty"`
		LiveSpeed    *int64   `json:"live_speed,omitempty"`
		LivePeers    *int     `json:"live_peers,omitempty"`
		LiveETA      *int64   `json:"live_eta,omitempty"`
	}

	var response []DownloadResponse
	for _, d := range downloads {
		dr := DownloadResponse{Download: d}
		if live, ok := activeMap[d.ID]; ok {
			dr.LiveProgress = &live.Progress
			dr.LiveSpeed = &live.Speed
			dr.LivePeers = &live.Peers
			dr.LiveETA = &live.ETA
		}
		response = append(response, dr)
	}

	return utils.SuccessData(c, response)
}

// HandleCancelDownload cancels an active download
func HandleCancelDownload(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.BadRequest(c, "Invalid download ID")
	}

	var download models.Download
	if err := config.DB.First(&download, id).Error; err != nil {
		return utils.NotFound(c, "Download not found")
	}

	if !download.IsActive() {
		return utils.BadRequest(c, "Download is not active")
	}

	// Cancel in torrent engine
	engine := enginepkg.GetEngine()
	_ = engine.CancelDownload(download.InfoHash)

	// Update DB
	config.DB.Model(&download).Updates(map[string]interface{}{
		"status": models.StatusCancelled,
	})

	log.Printf("🛑 Download cancelled: %s", download.Title)
	return utils.Success(c, "Download cancelled", nil)
}

// HandleDownloadProgress provides real-time download progress via SSE
func HandleDownloadProgress(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Send initial active downloads
		engine := enginepkg.GetEngine()
		active := engine.GetActiveDownloads()
		for _, update := range active {
			data, _ := json.Marshal(update)
			fmt.Fprintf(w, "data: %s\n\n", data)
		}
		w.Flush()

		// Register for live updates
		updates := make(chan enginepkg.ProgressUpdate, 50)
		engine.OnProgress(func(update enginepkg.ProgressUpdate) {
			select {
			case updates <- update:
			default:
				// Channel full, skip update
			}
		})

		// Stream updates
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case update := <-updates:
				data, _ := json.Marshal(update)
				_, err := fmt.Fprintf(w, "data: %s\n\n", data)
				if err != nil {
					return // Client disconnected
				}
				w.Flush()

			case <-ticker.C:
				// Send heartbeat to keep connection alive
				_, err := fmt.Fprintf(w, ": heartbeat\n\n")
				if err != nil {
					return // Client disconnected
				}
				w.Flush()
			}
		}
	})

	return nil
}

// ── Helpers ─────────────────────────────────────────────────────────

// extractInfoHash extracts the info hash from a magnet URI
func extractInfoHash(magnetURI string) string {
	// Look for xt=urn:btih:HASH
	parts := strings.Split(magnetURI, "&")
	for _, part := range parts {
		part = strings.TrimPrefix(part, "magnet:?")
		if strings.HasPrefix(part, "xt=urn:btih:") {
			hash := strings.TrimPrefix(part, "xt=urn:btih:")
			// Hash can be hex (40 chars) or base32 (32 chars)
			if len(hash) >= 32 {
				return strings.ToLower(hash[:40]) // Normalize to lowercase hex
			}
		}
	}
	return ""
}
