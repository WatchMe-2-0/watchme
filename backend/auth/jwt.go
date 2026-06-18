package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"watchme/config"
)

// Claims represents the JWT token claims
type Claims struct {
	ProfileID uint   `json:"profile_id"`
	UserID    uint   `json:"user_id"`
	Role      string `json:"role"`
	IsKids    bool   `json:"is_kids"`
	jwt.RegisteredClaims
}

// GenerateToken creates a JWT token for an authenticated profile
func GenerateToken(profileID uint, userID uint, role string, isKids bool) (string, time.Time, error) {
	cfg := config.Get()
	if cfg.JWTSecret == "" {
		return "", time.Time{}, fmt.Errorf("JWT secret not configured")
	}

	expiresAt := time.Now().Add(time.Duration(cfg.SessionExpiry) * 24 * time.Hour)

	claims := &Claims{
		ProfileID: profileID,
		UserID:    userID,
		Role:      role,
		IsKids:    isKids,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "watchme",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken parses and validates a JWT token, returning the claims
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.Get()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
