package routes

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"backend/config"
	"backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
)

// Upload movie file to MinIO
func UploadMovie(c *fiber.Ctx) error {
	// Get title from form
	title := c.FormValue("title")
	if title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Movie title is required"})
	}

	// Get movie file
	movieFile, err := c.FormFile("movie")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Movie file is required"})
	}

	// Get poster file
	posterFile, err := c.FormFile("poster")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Poster image is required"})
	}

	// Open movie file
	movieSrc, err := movieFile.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot open movie file"})
	}
	defer movieSrc.Close()

	// Open poster file
	posterSrc, err := posterFile.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot open poster file"})
	}
	defer posterSrc.Close()

	// Generate unique filenames
	movieObject := fmt.Sprintf("%d-%s", time.Now().Unix(), movieFile.Filename)
	posterObject := fmt.Sprintf("%d-%s", time.Now().Unix(), posterFile.Filename)

	// Upload movie to MinIO
	_, err = config.MinioClient.PutObject(
		context.Background(),
		"movies",
		movieObject,
		movieSrc,
		movieFile.Size,
		minio.PutObjectOptions{ContentType: movieFile.Header.Get("Content-Type")},
	)
	if err != nil {
		log.Printf("❌ Failed to upload movie: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to upload movie"})
	}

	log.Println("🎬 Movie uploaded to MinIO:", movieObject)

	// Upload poster to MinIO
	_, err = config.MinioClient.PutObject(
		context.Background(),
		"posters",
		posterObject,
		posterSrc,
		posterFile.Size,
		minio.PutObjectOptions{ContentType: posterFile.Header.Get("Content-Type")},
	)
	if err != nil {
		log.Printf("❌ Failed to upload poster: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to upload poster"})
	}

	log.Println("🖼️ Poster uploaded to MinIO:", posterObject)

	// Create URLs
	movieURL := fmt.Sprintf("http://localhost:9000/movies/%s", movieObject)
	posterURL := fmt.Sprintf("http://localhost:9000/posters/%s", posterObject)

	// Save movie metadata in PostgreSQL
	movie := models.Movie{
		Title:     title,
		PosterURL: posterURL,
		StreamURL: movieURL,
	}
	result := config.DB.Create(&movie)
	if result.Error != nil {
		log.Printf("❌ Failed to save movie metadata: %v", result.Error)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save movie metadata"})
	}

	log.Println("✅ Metadata saved in PostgreSQL:", title)

	// Return response
	return c.JSON(fiber.Map{
		"message":   "Upload successful",
		"title":     title,
		"posterURL": posterURL,
		"streamURL": movieURL,
	})
}

func ListMovies(c *fiber.Ctx) error {
	// List all objects in the "movies" bucket
	objectCh := config.MinioClient.ListObjects(
		context.Background(),
		"movies",
		minio.ListObjectsOptions{},
	)

	// Store movie URLs
	var movies []string

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("❌ Failed to list object: %v", object.Err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to list movies"})
		}

		// Generate the public URL for each movie
		movieURL := fmt.Sprintf("http://localhost:9000/movies/%s", object.Key)
		movies = append(movies, movieURL)
	}

	// Return the list of movies
	return c.JSON(fiber.Map{"movies": movies})
}

func StreamMovie(c *fiber.Ctx) error {
	movieName := c.Params("name")
	if movieName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Movie name is required"})
	}

	// Generate a pre-signed URL for streaming
	reqParams := make(url.Values)
	reqParams.Set("response-content-type", "video/mp4")

	presignedURL, err := config.MinioClient.PresignedGetObject(
		context.Background(),
		"movies",
		movieName,
		time.Minute*10, // URL expires in 10 minutes
		reqParams,
	)
	if err != nil {
		log.Printf("❌ Failed to generate streaming URL: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate streaming URL"})
	}

	// Redirect user to the streaming URL
	return c.Redirect(presignedURL.String(), 302)
}
