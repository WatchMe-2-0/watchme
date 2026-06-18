package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"watchme/config"
	"watchme/tmdb"
	"watchme/utils"
)

// HandleTMDBSearch searches movies via TMDB
func HandleTMDBSearch(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return utils.BadRequest(c, "Search query is required")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	client := tmdb.GetClient()
	result, err := client.SearchMovies(query, page)
	if err != nil {
		return utils.InternalError(c, fmt.Sprintf("TMDB search failed: %v", err))
	}

	return utils.SuccessData(c, result)
}

// HandleTMDBTrending returns trending movies
func HandleTMDBTrending(c *fiber.Ctx) error {
	timeWindow := c.Query("window", "week") // "day" or "week"
	page, _ := strconv.Atoi(c.Query("page", "1"))

	client := tmdb.GetClient()
	result, err := client.Trending(timeWindow, page)
	if err != nil {
		return utils.InternalError(c, fmt.Sprintf("TMDB trending failed: %v", err))
	}

	return utils.SuccessData(c, result)
}

// HandleTMDBTopRated returns top-rated movies
func HandleTMDBTopRated(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))

	client := tmdb.GetClient()
	result, err := client.TopRated(page)
	if err != nil {
		return utils.InternalError(c, fmt.Sprintf("TMDB top rated failed: %v", err))
	}

	return utils.SuccessData(c, result)
}

// HandleTMDBByGenre returns movies filtered by genre
func HandleTMDBByGenre(c *fiber.Ctx) error {
	genreID, err := c.ParamsInt("id")
	if err != nil {
		return utils.BadRequest(c, "Invalid genre ID")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))

	client := tmdb.GetClient()
	result, err := client.ByGenre(genreID, page)
	if err != nil {
		return utils.InternalError(c, fmt.Sprintf("TMDB genre browse failed: %v", err))
	}

	return utils.SuccessData(c, result)
}

// HandleTMDBMovieDetail returns detailed info for a single movie
func HandleTMDBMovieDetail(c *fiber.Ctx) error {
	movieID, err := c.ParamsInt("id")
	if err != nil {
		return utils.BadRequest(c, "Invalid movie ID")
	}

	client := tmdb.GetClient()
	detail, err := client.GetMovieDetail(movieID)
	if err != nil {
		return utils.InternalError(c, fmt.Sprintf("TMDB movie detail failed: %v", err))
	}

	// Also fetch certification
	cert, _ := client.GetCertification(movieID)

	return utils.SuccessData(c, fiber.Map{
		"movie":         detail,
		"certification": cert,
	})
}

// HandleTMDBGenres returns the full genre list
func HandleTMDBGenres(c *fiber.Ctx) error {
	client := tmdb.GetClient()
	genres, err := client.GetGenres()
	if err != nil {
		return utils.InternalError(c, fmt.Sprintf("TMDB genres failed: %v", err))
	}

	return utils.SuccessData(c, genres)
}

// HandleTMDBPosterProxy proxies poster images from TMDB with caching
func HandleTMDBPosterProxy(c *fiber.Ctx) error {
	path := c.Params("*")
	if path == "" {
		return utils.BadRequest(c, "Poster path is required")
	}

	imageURL := fmt.Sprintf("https://image.tmdb.org/t/p/w500/%s", path)

	// Check if we have it cached locally
	cfg := config.Get()
	localPath := cfg.PosterDir + "/tmdb-" + path

	// Try serving from local cache first
	if c.SendFile(localPath) == nil {
		c.Set("Cache-Control", "public, max-age=604800") // 7 days
		return nil
	}

	// Fetch from TMDB
	resp, err := http.Get(imageURL)
	if err != nil {
		return utils.InternalError(c, "Failed to fetch poster from TMDB")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return utils.NotFound(c, "Poster not found on TMDB")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return utils.InternalError(c, "Failed to read poster data")
	}

	// Cache locally (fire and forget)
	go func() {
		_ = saveToFile(localPath, body)
	}()

	// Set cache headers
	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	c.Set("Cache-Control", "public, max-age=604800") // 7 days

	return c.Send(body)
}

// saveToFile writes data to a file, creating directories as needed
func saveToFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// HandleTMDBPosterProxySized proxies poster images at specific sizes
func HandleTMDBPosterProxySized(c *fiber.Ctx) error {
	size := c.Params("size", "w500") // w92, w154, w185, w342, w500, w780, original
	path := c.Params("*")

	validSizes := map[string]bool{
		"w92": true, "w154": true, "w185": true,
		"w342": true, "w500": true, "w780": true, "original": true,
	}
	if !validSizes[size] {
		size = "w500"
	}

	imageURL := fmt.Sprintf("https://image.tmdb.org/t/p/%s/%s", size, path)

	resp, err := http.Get(imageURL)
	if err != nil {
		return utils.InternalError(c, "Failed to fetch poster")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return utils.InternalError(c, "Failed to read poster")
	}

	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(7*24*time.Hour/time.Second)))

	return c.Send(body)
}
