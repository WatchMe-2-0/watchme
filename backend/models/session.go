package models

import (
	"time"
)

// Session represents an authenticated profile session
type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProfileID uint      `gorm:"index;not null" json:"profile_id"`
	Token     string    `gorm:"type:varchar(500);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Profile Profile `gorm:"foreignKey:ProfileID" json:"-"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
