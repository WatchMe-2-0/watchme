package models

import (
	"time"
)

// DownloadStatus represents the state of a torrent download
type DownloadStatus string

const (
	StatusQueued      DownloadStatus = "queued"
	StatusDownloading DownloadStatus = "downloading"
	StatusSeeding     DownloadStatus = "seeding"
	StatusCompleted   DownloadStatus = "completed"
	StatusFailed      DownloadStatus = "failed"
	StatusCancelled   DownloadStatus = "cancelled"
)

// Download represents a torrent download job
type Download struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	ProfileID  uint           `gorm:"index;not null" json:"profile_id"`
	MagnetLink string         `gorm:"type:text;not null" json:"magnet_link"`
	InfoHash   string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"info_hash"`
	Title      string         `gorm:"type:varchar(255)" json:"title"`
	Status     DownloadStatus `gorm:"type:varchar(20);not null;default:'queued'" json:"status"`
	Progress   float64        `gorm:"type:decimal(5,2);default:0" json:"progress"`   // 0.00 to 100.00
	Speed      int64          `json:"speed"`                                          // Bytes per second
	ETA        int64          `json:"eta"`                                            // Seconds remaining
	Peers      int            `json:"peers"`                                          // Connected peers
	FilePath   string         `gorm:"type:varchar(500)" json:"file_path,omitempty"`   // Final file path once complete
	FileSize   int64          `json:"file_size"`                                      // Total bytes
	Downloaded int64          `json:"downloaded"`                                     // Bytes downloaded so far
	Error      string         `gorm:"type:text" json:"error,omitempty"`               // Error message if failed
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Profile Profile `gorm:"foreignKey:ProfileID" json:"-"`
}

// IsActive checks if the download is still in progress
func (d *Download) IsActive() bool {
	return d.Status == StatusQueued || d.Status == StatusDownloading
}
