package torrent

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"watchme/config"
)

// Engine wraps the anacrolix/torrent client with WATCHME-specific logic
type Engine struct {
	client    *torrent.Client
	mu        sync.RWMutex
	active    map[string]*TorrentJob // keyed by info hash
	listeners []ProgressListener
}

// TorrentJob represents an active torrent download
type TorrentJob struct {
	InfoHash   string
	Title      string
	Torrent    *torrent.Torrent
	ProfileID  uint
	StartedAt  time.Time
	DownloadID uint // DB record ID
}

// ProgressListener receives progress updates
type ProgressListener func(update ProgressUpdate)

// ProgressUpdate contains download progress information
type ProgressUpdate struct {
	InfoHash   string  `json:"info_hash"`
	DownloadID uint    `json:"download_id"`
	Title      string  `json:"title"`
	Progress   float64 `json:"progress"`
	Speed      int64   `json:"speed"`       // bytes/sec
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Peers      int     `json:"peers"`
	Status     string  `json:"status"`
	ETA        int64   `json:"eta"`         // seconds remaining
	FilePath   string  `json:"file_path,omitempty"`
}

var (
	engineInstance *Engine
	engineOnce     sync.Once
)

// GetEngine returns the singleton torrent engine
func GetEngine() *Engine {
	engineOnce.Do(func() {
		engineInstance = &Engine{
			active: make(map[string]*TorrentJob),
		}
	})
	return engineInstance
}

// Start initializes the torrent client
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client != nil {
		return nil // Already started
	}

	cfg := config.Get()

	// Ensure download directory exists
	downloadDir := cfg.DownloadDir
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	// Configure torrent client for performance
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = downloadDir
	clientConfig.DefaultStorage = storage.NewFileByInfoHash(downloadDir)
	clientConfig.Seed = false        // Don't seed after download (local use)
	clientConfig.NoUpload = true     // Don't upload (save bandwidth)
	clientConfig.DisableIPv6 = false
	clientConfig.ListenPort = 0      // Random port

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create torrent client: %w", err)
	}

	e.client = client
	log.Println("🌊 Torrent engine started")

	return nil
}

// Stop shuts down the torrent client gracefully
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client != nil {
		e.client.Close()
		e.client = nil
		log.Println("🛑 Torrent engine stopped")
	}
}

// AddMagnet starts downloading a magnet link
func (e *Engine) AddMagnet(magnetURI string, profileID uint, downloadID uint, title string) (*TorrentJob, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.client == nil {
		return nil, fmt.Errorf("torrent engine not started")
	}

	t, err := e.client.AddMagnet(magnetURI)
	if err != nil {
		return nil, fmt.Errorf("failed to add magnet link: %w", err)
	}

	infoHash := t.InfoHash().HexString()

	// Check if already downloading
	if job, exists := e.active[infoHash]; exists {
		return job, nil
	}

	job := &TorrentJob{
		InfoHash:   infoHash,
		Title:      title,
		Torrent:    t,
		ProfileID:  profileID,
		StartedAt:  time.Now(),
		DownloadID: downloadID,
	}

	e.active[infoHash] = job

	// Start download in goroutine
	go e.downloadWorker(job)

	log.Printf("🧲 Magnet added: %s (hash: %s)", title, infoHash[:12])
	return job, nil
}

// downloadWorker manages a single torrent download
func (e *Engine) downloadWorker(job *TorrentJob) {
	t := job.Torrent

	// Wait for torrent metadata (info)
	log.Printf("⏳ Waiting for metadata: %s", job.InfoHash[:12])
	<-t.GotInfo()

	// Update title from torrent metadata if not set
	if job.Title == "" {
		job.Title = t.Info().Name
	}

	// Enable sequential download for streaming support
	t.DownloadAll()

	// Get the largest file (the movie)
	var largestFile *torrent.File
	for _, f := range t.Files() {
		if largestFile == nil || f.Length() > largestFile.Length() {
			largestFile = f
		}
	}

	if largestFile != nil {
		// Prioritize first and last pieces for streaming
		largestFile.SetPriority(torrent.PiecePriorityNormal)
	}

	// Monitor progress
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastDownloaded int64
	lastCheck := time.Now()

	for range ticker.C {
		e.mu.RLock()
		_, stillActive := e.active[job.InfoHash]
		e.mu.RUnlock()

		if !stillActive {
			return // Cancelled
		}

		stats := t.Stats()
		bytesCompleted := t.BytesCompleted()
		totalBytes := t.Length()

		// Calculate speed
		now := time.Now()
		elapsed := now.Sub(lastCheck).Seconds()
		speed := int64(0)
		if elapsed > 0 {
			speed = int64(float64(bytesCompleted-lastDownloaded) / elapsed)
		}
		lastDownloaded = bytesCompleted
		lastCheck = now

		// Calculate ETA
		var eta int64
		if speed > 0 {
			remaining := totalBytes - bytesCompleted
			eta = remaining / speed
		}

		// Calculate progress
		progress := float64(0)
		if totalBytes > 0 {
			progress = float64(bytesCompleted) / float64(totalBytes) * 100
		}

		// Determine file path
		filePath := ""
		if largestFile != nil {
			cfg := config.Get()
			filePath = filepath.Join(cfg.DownloadDir, largestFile.Path())
		}

		update := ProgressUpdate{
			InfoHash:   job.InfoHash,
			DownloadID: job.DownloadID,
			Title:      job.Title,
			Progress:   progress,
			Speed:      speed,
			Downloaded: bytesCompleted,
			Total:      totalBytes,
			Peers:      stats.ActivePeers,
			Status:     "downloading",
			ETA:        eta,
			FilePath:   filePath,
		}

		// Check completion
		if bytesCompleted >= totalBytes && totalBytes > 0 {
			update.Status = "completed"
			update.Progress = 100

			// Notify listeners
			e.notifyListeners(update)

			// Remove from active
			e.mu.Lock()
			delete(e.active, job.InfoHash)
			e.mu.Unlock()

			log.Printf("✅ Download complete: %s", job.Title)
			return
		}

		// Notify listeners
		e.notifyListeners(update)
	}
}

// CancelDownload stops and removes a torrent download
func (e *Engine) CancelDownload(infoHash string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	job, exists := e.active[infoHash]
	if !exists {
		return fmt.Errorf("download not found: %s", infoHash)
	}

	job.Torrent.Drop()
	delete(e.active, infoHash)

	log.Printf("🛑 Download cancelled: %s", infoHash[:12])
	return nil
}

// GetActiveDownloads returns all currently active downloads
func (e *Engine) GetActiveDownloads() []ProgressUpdate {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var updates []ProgressUpdate
	for _, job := range e.active {
		t := job.Torrent
		bytesCompleted := t.BytesCompleted()
		totalBytes := t.Length()

		progress := float64(0)
		if totalBytes > 0 {
			progress = float64(bytesCompleted) / float64(totalBytes) * 100
		}

		updates = append(updates, ProgressUpdate{
			InfoHash:   job.InfoHash,
			DownloadID: job.DownloadID,
			Title:      job.Title,
			Progress:   progress,
			Downloaded: bytesCompleted,
			Total:      totalBytes,
			Peers:      t.Stats().ActivePeers,
			Status:     "downloading",
		})
	}

	return updates
}

// GetFilePath returns the file path for a completed or in-progress torrent
func (e *Engine) GetFilePath(infoHash string) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	job, exists := e.active[infoHash]
	if !exists {
		return "", fmt.Errorf("download not found")
	}

	t := job.Torrent
	select {
	case <-t.GotInfo():
		// Has info, find largest file
		var largestFile *torrent.File
		for _, f := range t.Files() {
			if largestFile == nil || f.Length() > largestFile.Length() {
				largestFile = f
			}
		}
		if largestFile != nil {
			cfg := config.Get()
			return filepath.Join(cfg.DownloadDir, largestFile.Path()), nil
		}
		return "", fmt.Errorf("no files in torrent")
	default:
		return "", fmt.Errorf("waiting for metadata")
	}
}

// OnProgress registers a listener for download progress updates
func (e *Engine) OnProgress(listener ProgressListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, listener)
}

func (e *Engine) notifyListeners(update ProgressUpdate) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, listener := range e.listeners {
		listener(update)
	}
}
