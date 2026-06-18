package utils

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse is the standard response format
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success sends a successful JSON response
func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessData sends a successful JSON response with only data (no message)
func SuccessData(c *fiber.Ctx, data interface{}) error {
	return c.JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// Error sends an error JSON response with the given status code
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}

// BadRequest sends a 400 error
func BadRequest(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusBadRequest, message)
}

// Unauthorized sends a 401 error
func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusUnauthorized, message)
}

// Forbidden sends a 403 error
func Forbidden(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusForbidden, message)
}

// NotFound sends a 404 error
func NotFound(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusNotFound, message)
}

// InternalError sends a 500 error
func InternalError(c *fiber.Ctx, message string) error {
	return Error(c, fiber.StatusInternalServerError, message)
}
