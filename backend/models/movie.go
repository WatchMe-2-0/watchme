package models

import (
	"time"
)

// Movie represents a movie stored locally
type Movie struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Title         string    `gorm:"type:varchar(255);not null" json:"title"`
	TmdbID        int       `gorm:"index" json:"tmdb_id,omitempty"`
	Rating        float64   `gorm:"type:decimal(3,1);default:0" json:"rating"`
	Certification string    `gorm:"type:varchar(10)" json:"certification,omitempty"` // G, PG, PG-13, R, NC-17
	Year          int       `json:"year,omitempty"`
	Genre         string    `gorm:"type:varchar(255)" json:"genre,omitempty"`    // Comma-separated genre names
	Overview      string    `gorm:"type:text" json:"overview,omitempty"`
	PosterPath    string    `gorm:"type:varchar(500)" json:"poster_path,omitempty"`  // Local path to poster file
	BackdropPath  string    `gorm:"type:varchar(500)" json:"backdrop_path,omitempty"` // Local path to backdrop
	Runtime       int       `json:"runtime,omitempty"`                                // Minutes
	FilePath      string    `gorm:"type:varchar(500);not null" json:"file_path"`      // Local path to movie file
	FileSize      int64     `json:"file_size"`                                        // Bytes
	Source        string    `gorm:"type:varchar(20);not null;default:'upload'" json:"source"` // "upload" or "torrent"
	InfoHash      string    `gorm:"type:varchar(64);index" json:"info_hash,omitempty"`        // Torrent info hash
	ProfileID     uint      `gorm:"index" json:"profile_id,omitempty"`                        // Who added it
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// IsKidsSafe checks if movie is appropriate for kids profiles
func (m *Movie) IsKidsSafe() bool {
	safe := map[string]bool{"G": true, "PG": true, "": true}
	return safe[m.Certification]
}
