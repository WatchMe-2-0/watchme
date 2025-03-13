package main

import (
	"fmt"
	"log"

	"backend/config"
	"backend/routes"
	"backend/secrets"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	secrets.LoadConfig()

	config.InitMinio()
	config.ConnectDatabase()

	app := fiber.New(fiber.Config{
		BodyLimit: 5 * 1024 * 1024 * 1024, // 5GB upload limit
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:  "http://localhost:3000",
		AllowMethods:  "GET,POST,PUT,DELETE,OPTIONS,HEAD,PATCH",
		AllowHeaders:  "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders: "Content-Length,Content-Type",
	}))

	//cors
	app.Post("/upload", routes.UploadMovie)
	app.Get("/movies", routes.ListMovies)
	app.Get("/stream/:name", routes.StreamMovie)
	app.Delete("/movies/:id", routes.DeleteMovie)
	app.Get("/posters/:name", routes.GetPoster)

	fmt.Println("🚀 Home Network API is running on http://localhost:8000")
	log.Fatal(app.Listen(":8000"))
}
