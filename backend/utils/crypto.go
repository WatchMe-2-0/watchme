package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword verifies a password against a bcrypt hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPIN hashes a numeric PIN using bcrypt
func HashPIN(pin string) (string, error) {
	return HashPassword(pin)
}

// CheckPIN verifies a PIN against a bcrypt hash
func CheckPIN(pin, hash string) bool {
	return CheckPassword(pin, hash)
}

// GenerateRecoveryKey generates a cryptographically secure recovery key
// Returns the hex-encoded key (64 characters for 32-byte key)
func GenerateRecoveryKey(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate recovery key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateJWTSecret generates a random secret for JWT signing
func GenerateJWTSecret() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
