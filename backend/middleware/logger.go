package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger is a custom request logging middleware
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log after response
		duration := time.Since(start)
		status := c.Response().StatusCode()

		logFn := log.Printf
		if status >= 500 {
			logFn("❌ %s %s → %d (%s)", c.Method(), c.Path(), status, duration)
		} else if status >= 400 {
			logFn("⚠️  %s %s → %d (%s)", c.Method(), c.Path(), status, duration)
		} else {
			logFn("✅ %s %s → %d (%s)", c.Method(), c.Path(), status, duration)
		}

		return err
	}
}
