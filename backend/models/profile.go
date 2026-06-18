package models

import (
	"time"
)

// Profile represents a user profile (Netflix-style)
type Profile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Avatar    string    `gorm:"type:varchar(50);not null;default:'aurora'" json:"avatar"` // Preset avatar name
	PINHash   string    `gorm:"type:varchar(255);not null" json:"-"`
	IsKids    bool      `gorm:"default:false" json:"is_kids"`
	Port      int       `gorm:"default:0" json:"port,omitempty"` // Custom port binding (0 = none)
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Sessions []Session `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE" json:"-"`
}
