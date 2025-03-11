package models

import "time"

type Movie struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"type:varchar(255);not null"`
	PosterURL string    `gorm:"type:varchar(255);not null"` // Stores the MinIO poster URL
	StreamURL string    `gorm:"type:varchar(255);not null"` // Stores the MinIO movie URL
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
