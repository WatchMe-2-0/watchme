package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// SetupCORS configures CORS for the API
func SetupCORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "http://localhost:2103, http://localhost:3000, http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,HEAD,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Content-Type, Content-Range, Accept-Ranges",
		AllowCredentials: true,
	})
}
