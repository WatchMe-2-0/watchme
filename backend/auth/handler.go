package auth

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"watchme/config"
	"watchme/models"
	"watchme/utils"
)

// ── Request/Response Types ──────────────────────────────────────────

// SetupRequest is the admin first-time setup payload
type SetupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SetupResponse is returned after admin creation
type SetupResponse struct {
	Message     string `json:"message"`
	RecoveryKey string `json:"recovery_key"`
	Username    string `json:"username"`
}

// LoginRequest is the profile PIN login payload
type LoginRequest struct {
	ProfileID uint   `json:"profile_id"`
	PIN       string `json:"pin"`
}

// RecoveryRequest is the password recovery payload
type RecoveryRequest struct {
	RecoveryKey string `json:"recovery_key"`
	NewPassword string `json:"new_password"`
}

// CreateProfileRequest is the payload for creating a new profile
type CreateProfileRequest struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	PIN    string `json:"pin"`
	IsKids bool   `json:"is_kids"`
	Port   int    `json:"port,omitempty"`
}

// UpdateProfileRequest is the payload for updating a profile
type UpdateProfileRequest struct {
	Name   *string `json:"name,omitempty"`
	Avatar *string `json:"avatar,omitempty"`
	PIN    *string `json:"pin,omitempty"`
	IsKids *bool   `json:"is_kids,omitempty"`
	Port   *int    `json:"port,omitempty"`
}

// ── Auth Handlers ───────────────────────────────────────────────────

// HandleAuthStatus checks if admin setup is complete
func HandleAuthStatus(c *fiber.Ctx) error {
	var count int64
	config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	return utils.SuccessData(c, fiber.Map{
		"setup_complete": count > 0,
	})
}

// HandleSetup creates the first admin account
func HandleSetup(c *fiber.Ctx) error {
	// Check if admin already exists
	var count int64
	config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return utils.BadRequest(c, "Admin account already exists")
	}

	var req SetupRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	// Validate
	if req.Username == "" {
		return utils.BadRequest(c, "Username is required")
	}
	if len(req.Password) < config.AdminPasswordLength {
		return utils.BadRequest(c, "Password must be at least 6 characters")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("❌ Failed to hash password: %v", err)
		return utils.InternalError(c, "Failed to create admin account")
	}

	// Generate recovery key
	recoveryKey, err := utils.GenerateRecoveryKey(config.RecoveryKeyLength)
	if err != nil {
		log.Printf("❌ Failed to generate recovery key: %v", err)
		return utils.InternalError(c, "Failed to generate recovery key")
	}

	recoveryKeyHash, err := utils.HashPassword(recoveryKey)
	if err != nil {
		log.Printf("❌ Failed to hash recovery key: %v", err)
		return utils.InternalError(c, "Failed to create admin account")
	}

	// Create admin user
	admin := models.User{
		Username:        req.Username,
		PasswordHash:    passwordHash,
		RecoveryKeyHash: recoveryKeyHash,
		Role:            "admin",
	}

	if err := config.DB.Create(&admin).Error; err != nil {
		log.Printf("❌ Failed to create admin: %v", err)
		return utils.InternalError(c, "Failed to create admin account")
	}

	// Create default admin profile
	defaultPINHash, _ := utils.HashPIN("0000")
	defaultProfile := models.Profile{
		UserID:  admin.ID,
		Name:    req.Username,
		Avatar:  "aurora",
		PINHash: defaultPINHash,
		IsKids:  false,
	}

	if err := config.DB.Create(&defaultProfile).Error; err != nil {
		log.Printf("❌ Failed to create default profile: %v", err)
		return utils.InternalError(c, "Failed to create default profile")
	}

	log.Printf("✅ Admin account created: %s", req.Username)

	return utils.Success(c, "Admin account created successfully", SetupResponse{
		Message:     "Admin account created. Save your recovery key — it will not be shown again.",
		RecoveryKey: recoveryKey,
		Username:    req.Username,
	})
}

// HandleLogin authenticates a profile via PIN and returns JWT
func HandleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	if req.ProfileID == 0 {
		return utils.BadRequest(c, "Profile ID is required")
	}
	if req.PIN == "" {
		return utils.BadRequest(c, "PIN is required")
	}

	// Find profile
	var profile models.Profile
	if err := config.DB.Preload("User").First(&profile, req.ProfileID).Error; err != nil {
		return utils.Unauthorized(c, "Invalid profile or PIN")
	}

	// Verify PIN
	if !utils.CheckPIN(req.PIN, profile.PINHash) {
		return utils.Unauthorized(c, "Invalid profile or PIN")
	}

	// Generate JWT
	token, expiresAt, err := GenerateToken(
		profile.ID,
		profile.UserID,
		profile.User.Role,
		profile.IsKids,
	)
	if err != nil {
		log.Printf("❌ Failed to generate token: %v", err)
		return utils.InternalError(c, "Failed to create session")
	}

	// Store session in DB
	session := models.Session{
		ProfileID: profile.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	config.DB.Create(&session)

	// Set httpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "watchme_token",
		Value:    token,
		Expires:  expiresAt,
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	return utils.Success(c, "Login successful", fiber.Map{
		"token":   token,
		"profile": profile,
		"expires": expiresAt,
	})
}

// HandleLogout clears the session
func HandleLogout(c *fiber.Ctx) error {
	// Clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "watchme_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	// Remove session from DB if authenticated
	tokenString := c.Cookies("watchme_token")
	if tokenString == "" {
		tokenString = c.Get("Authorization")
		if len(tokenString) > 7 {
			tokenString = tokenString[7:]
		}
	}
	if tokenString != "" {
		config.DB.Where("token = ?", tokenString).Delete(&models.Session{})
	}

	return utils.Success(c, "Logged out successfully", nil)
}

// HandleRecovery allows password reset via recovery key
func HandleRecovery(c *fiber.Ctx) error {
	var req RecoveryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	if req.RecoveryKey == "" {
		return utils.BadRequest(c, "Recovery key is required")
	}
	if len(req.NewPassword) < config.AdminPasswordLength {
		return utils.BadRequest(c, "New password must be at least 6 characters")
	}

	// Find admin
	var admin models.User
	if err := config.DB.Where("role = ?", "admin").First(&admin).Error; err != nil {
		return utils.NotFound(c, "Admin account not found")
	}

	// Verify recovery key
	if !utils.CheckPassword(req.RecoveryKey, admin.RecoveryKeyHash) {
		return utils.Unauthorized(c, "Invalid recovery key")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.InternalError(c, "Failed to update password")
	}

	// Update password
	config.DB.Model(&admin).Update("password_hash", newPasswordHash)

	log.Printf("✅ Admin password recovered for: %s", admin.Username)

	return utils.Success(c, "Password updated successfully", nil)
}

// HandleChangePassword allows admin to change their password
func HandleChangePassword(c *fiber.Ctx) error {
	claims := GetClaims(c)
	if claims == nil || claims.Role != "admin" {
		return utils.Forbidden(c, "Admin access required")
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	if len(req.NewPassword) < config.AdminPasswordLength {
		return utils.BadRequest(c, "New password must be at least 6 characters")
	}

	// Find admin
	var admin models.User
	if err := config.DB.First(&admin, claims.UserID).Error; err != nil {
		return utils.NotFound(c, "User not found")
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, admin.PasswordHash) {
		return utils.Unauthorized(c, "Current password is incorrect")
	}

	// Hash new password
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.InternalError(c, "Failed to update password")
	}

	config.DB.Model(&admin).Update("password_hash", newHash)

	return utils.Success(c, "Password changed successfully", nil)
}

// ── Profile Handlers ────────────────────────────────────────────────

// HandleListProfiles returns all profiles (public, for profile selection screen)
func HandleListProfiles(c *fiber.Ctx) error {
	var profiles []models.Profile
	config.DB.Select("id, name, avatar, is_kids, port").Find(&profiles)
	return utils.SuccessData(c, profiles)
}

// HandleCreateProfile creates a new profile (admin only)
func HandleCreateProfile(c *fiber.Ctx) error {
	var req CreateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	if req.Name == "" {
		return utils.BadRequest(c, "Profile name is required")
	}
	if len(req.PIN) < config.MinPINLength || len(req.PIN) > config.MaxPINLength {
		return utils.BadRequest(c, "PIN must be 3 or 4 digits")
	}

	// Validate avatar
	validAvatar := false
	for _, preset := range config.AvatarPresets {
		if req.Avatar == preset {
			validAvatar = true
			break
		}
	}
	if !validAvatar {
		req.Avatar = "aurora" // Default avatar
	}

	// Get admin user ID
	claims := GetClaims(c)
	if claims == nil {
		return utils.Unauthorized(c, "Authentication required")
	}

	// Hash PIN
	pinHash, err := utils.HashPIN(req.PIN)
	if err != nil {
		return utils.InternalError(c, "Failed to create profile")
	}

	profile := models.Profile{
		UserID:  claims.UserID,
		Name:    req.Name,
		Avatar:  req.Avatar,
		PINHash: pinHash,
		IsKids:  req.IsKids,
		Port:    req.Port,
	}

	if err := config.DB.Create(&profile).Error; err != nil {
		log.Printf("❌ Failed to create profile: %v", err)
		return utils.InternalError(c, "Failed to create profile")
	}

	log.Printf("✅ Profile created: %s", req.Name)

	return utils.Success(c, "Profile created", fiber.Map{
		"id":     profile.ID,
		"name":   profile.Name,
		"avatar": profile.Avatar,
		"is_kids": profile.IsKids,
	})
}

// HandleUpdateProfile updates an existing profile (admin only)
func HandleUpdateProfile(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.BadRequest(c, "Invalid profile ID")
	}

	var profile models.Profile
	if err := config.DB.First(&profile, id).Error; err != nil {
		return utils.NotFound(c, "Profile not found")
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequest(c, "Invalid request body")
	}

	// Apply updates
	if req.Name != nil {
		profile.Name = *req.Name
	}
	if req.Avatar != nil {
		profile.Avatar = *req.Avatar
	}
	if req.IsKids != nil {
		profile.IsKids = *req.IsKids
	}
	if req.Port != nil {
		profile.Port = *req.Port
	}
	if req.PIN != nil {
		if len(*req.PIN) < config.MinPINLength || len(*req.PIN) > config.MaxPINLength {
			return utils.BadRequest(c, "PIN must be 3 or 4 digits")
		}
		pinHash, err := utils.HashPIN(*req.PIN)
		if err != nil {
			return utils.InternalError(c, "Failed to update PIN")
		}
		profile.PINHash = pinHash
	}

	config.DB.Save(&profile)

	return utils.Success(c, "Profile updated", fiber.Map{
		"id":      profile.ID,
		"name":    profile.Name,
		"avatar":  profile.Avatar,
		"is_kids": profile.IsKids,
	})
}

// HandleDeleteProfile deletes a profile (admin only)
func HandleDeleteProfile(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.BadRequest(c, "Invalid profile ID")
	}

	var profile models.Profile
	if err := config.DB.First(&profile, id).Error; err != nil {
		return utils.NotFound(c, "Profile not found")
	}

	// Don't allow deleting the last profile
	var count int64
	config.DB.Model(&models.Profile{}).Count(&count)
	if count <= 1 {
		return utils.BadRequest(c, "Cannot delete the last profile")
	}

	config.DB.Delete(&profile)
	log.Printf("✅ Profile deleted: %s (ID: %d)", profile.Name, profile.ID)

	return utils.Success(c, "Profile deleted", nil)
}

// HandleGetAvatars returns available avatar presets
func HandleGetAvatars(c *fiber.Ctx) error {
	return utils.SuccessData(c, config.AvatarPresets)
}
