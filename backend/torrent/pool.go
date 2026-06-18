package torrent

import (
	"log"
	"sync"

	"watchme/config"
)

// Pool manages concurrent download slots using a worker pool pattern
type Pool struct {
	jobs     chan DownloadRequest
	wg       sync.WaitGroup
	maxSlots int
	engine   *Engine
}

// DownloadRequest represents a queued download
type DownloadRequest struct {
	MagnetURI  string
	ProfileID  uint
	DownloadID uint
	Title      string
}

var (
	poolInstance *Pool
	poolOnce     sync.Once
)

// GetPool returns the singleton download pool
func GetPool() *Pool {
	poolOnce.Do(func() {
		cfg := config.Get()
		poolInstance = &Pool{
			jobs:     make(chan DownloadRequest, 100), // Buffer up to 100 queued downloads
			maxSlots: cfg.MaxConcurrentDownloads,
			engine:   GetEngine(),
		}
	})
	return poolInstance
}

// Start launches the worker pool goroutines
func (p *Pool) Start() {
	for i := 0; i < p.maxSlots; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	log.Printf("🏊 Download pool started with %d workers", p.maxSlots)
}

// Stop gracefully shuts down the pool
func (p *Pool) Stop() {
	close(p.jobs)
	p.wg.Wait()
	log.Println("🛑 Download pool stopped")
}

// Submit adds a download request to the queue
func (p *Pool) Submit(req DownloadRequest) {
	p.jobs <- req
}

// worker processes download requests from the channel
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for req := range p.jobs {
		log.Printf("👷 Worker %d processing: %s", id, req.Title)

		_, err := p.engine.AddMagnet(req.MagnetURI, req.ProfileID, req.DownloadID, req.Title)
		if err != nil {
			log.Printf("❌ Worker %d failed to add magnet: %v", id, err)
			// Update download status in DB
			updateDownloadStatus(req.DownloadID, "failed", err.Error())
			continue
		}
	}
}

// updateDownloadStatus updates the download record in the database
func updateDownloadStatus(downloadID uint, status string, errMsg string) {
	if config.DB == nil {
		return
	}

	updates := map[string]interface{}{
		"status": status,
	}
	if errMsg != "" {
		updates["error"] = errMsg
	}

	config.DB.Table("downloads").Where("id = ?", downloadID).Updates(updates)
}
