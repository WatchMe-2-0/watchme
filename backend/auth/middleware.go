package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"watchme/config"
	"watchme/models"
	"watchme/utils"
)

// RequireAuth validates JWT token from cookie or Authorization header
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := extractToken(c)
		if tokenString == "" {
			return utils.Unauthorized(c, "Authentication required")
		}

		claims, err := ValidateToken(tokenString)
		if err != nil {
			return utils.Unauthorized(c, "Invalid or expired token")
		}

		// Verify profile still exists
		var profile models.Profile
		if err := config.DB.First(&profile, claims.ProfileID).Error; err != nil {
			return utils.Unauthorized(c, "Profile not found")
		}

		// Store claims in context for downstream handlers
		c.Locals("claims", claims)
		c.Locals("profile_id", claims.ProfileID)
		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("is_kids", claims.IsKids)

		return c.Next()
	}
}

// RequireAdmin ensures the authenticated user has admin role
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != "admin" {
			return utils.Forbidden(c, "Admin access required")
		}
		return c.Next()
	}
}

// KidsFilter middleware filters content for kids profiles
func KidsFilter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		isKids, ok := c.Locals("is_kids").(bool)
		if ok && isKids {
			c.Locals("kids_filter", true)
		}
		return c.Next()
	}
}

// extractToken gets the JWT token from cookie, Authorization header, or query parameter
func extractToken(c *fiber.Ctx) string {
	// Try cookie first
	token := c.Cookies("watchme_token")
	if token != "" {
		return token
	}

	// Fall back to Authorization header
	auth := c.Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Fall back to query parameter (needed for EventSource/SSE)
	token = c.Query("token")
	if token != "" {
		return token
	}

	return ""
}

// GetProfileID extracts the profile ID from the request context
func GetProfileID(c *fiber.Ctx) uint {
	id, ok := c.Locals("profile_id").(uint)
	if !ok {
		return 0
	}
	return id
}

// GetClaims extracts the full claims from the request context
func GetClaims(c *fiber.Ctx) *Claims {
	claims, ok := c.Locals("claims").(*Claims)
	if !ok {
		return nil
	}
	return claims
}
