package models

import (
	"time"
)

// User represents an admin or standard user account
type User struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Username        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	PasswordHash    string    `gorm:"type:varchar(255);not null" json:"-"`
	RecoveryKeyHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role            string    `gorm:"type:varchar(20);not null;default:'standard'" json:"role"` // "admin" or "standard"
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Profiles []Profile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"profiles,omitempty"`
}

// IsAdmin checks if the user has admin privileges
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}
