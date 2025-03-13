package routes

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"backend/config"
	"backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
)

// Upload movie file to MinIO
func UploadMovie(c *fiber.Ctx) error {
	// Get title string
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

	// Generate unique filenames with timestamp
	randomID := time.Now().Unix()
	movieObject := fmt.Sprintf("%d-%s", randomID, movieFile.Filename)
	posterObject := fmt.Sprintf("%d-%s", randomID, posterFile.Filename)

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
	posterURL := fmt.Sprintf("http://localhost:8000/posters/%s", posterObject)
	streamURL := fmt.Sprintf("http://localhost:8000/stream/%s", movieObject)

	// Save movie metadata in PostgreSQL
	movie := models.Movie{
		Title:     title,
		PosterURL: posterURL,
		StreamURL: streamURL,
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
		"streamURL": streamURL,
	})
}

func ListMovies(c *fiber.Ctx) error {
	// Fetch movie metadata from PostgreSQL using Prisma
	var movies []models.Movie
	result := config.DB.Find(&movies)

	if result.Error != nil {
		log.Printf("❌ Failed to fetch movies from DB: %v", result.Error)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch movies"})
	}

	// Return the movie list
	return c.JSON(movies)
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
		time.Minute*10,
		reqParams,
	)
	if err != nil {
		log.Printf("❌ Failed to generate streaming URL: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate streaming URL"})
	}

	return c.Redirect(presignedURL.String(), 302)
}

func DeleteMovie(c *fiber.Ctx) error {
	id := c.Params("id")

	// Find movie metadata in DB
	var movie models.Movie
	result := config.DB.First(&movie, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Movie not found"})
	}

	// Extract object names from URLs
	movieObject := getObjectName(movie.StreamURL)
	posterObject := getObjectName(movie.PosterURL)

	// Delete from MinIO storage
	err := config.MinioClient.RemoveObject(context.Background(), "movies", movieObject, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("❌ Failed to delete movie from MinIO: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete movie from storage"})
	}

	err = config.MinioClient.RemoveObject(context.Background(), "posters", posterObject, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("❌ Failed to delete poster from MinIO: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete poster from storage"})
	}

	// Delete from PostgreSQL
	config.DB.Delete(&movie)

	return c.JSON(fiber.Map{"message": "Movie deleted successfully"})
}

// Helper function to extract the object name from a URL
func getObjectName(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("❌ Failed to parse URL: %s", urlStr)
		return ""
	}
	parts := strings.Split(parsedURL.Path, "/")
	return parts[len(parts)-1]
}

func GetPoster(c *fiber.Ctx) error {
	posterName := c.Params("name")
	if posterName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Poster name is required"})
	}

	// Generate a pre-signed URL for poster access
	presignedURL, err := config.MinioClient.PresignedGetObject(
		context.Background(),
		"posters",
		posterName,
		time.Hour*1,
		nil,
	)
	if err != nil {
		log.Printf("❌ Failed to generate poster URL: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate poster URL"})
	}

	return c.Redirect(presignedURL.String(), 302)
}
