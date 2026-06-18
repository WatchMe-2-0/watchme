package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"watchme/auth"
	"watchme/config"
	"watchme/models"
	"watchme/utils"
)

// HandleListMovies returns all movies, filtered for kids profiles
func HandleListMovies(c *fiber.Ctx) error {
	var movies []models.Movie
	query := config.DB

	// Kids filter: only show safe content
	isKids, _ := c.Locals("kids_filter").(bool)
	if isKids {
		query = query.Where("certification IN ?", config.KidsSafeCertifications)
	}

	// Optional search
	search := c.Query("search")
	if search != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Movie{}).Count(&total)

	query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&movies)

	// Build response with poster URLs
	cfg := config.Get()
	type MovieResponse struct {
		models.Movie
		PosterURL  string `json:"poster_url"`
		StreamURL  string `json:"stream_url"`
	}

	var response []MovieResponse
	for _, m := range movies {
		mr := MovieResponse{Movie: m}
		if m.PosterPath != "" {
			mr.PosterURL = fmt.Sprintf("http://localhost:%s/api/posters/%s", cfg.ServerPort, filepath.Base(m.PosterPath))
		}
		mr.StreamURL = fmt.Sprintf("http://localhost:%s/api/stream/%d", cfg.ServerPort, m.ID)
		response = append(response, mr)
	}

	return utils.SuccessData(c, fiber.Map{
		"movies": response,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// HandleGetMovie returns a single movie's details
func HandleGetMovie(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.BadRequest(c, "Invalid movie ID")
	}

	var movie models.Movie
	if err := config.DB.First(&movie, id).Error; err != nil {
		return utils.NotFound(c, "Movie not found")
	}

	// Kids filter
	isKids, _ := c.Locals("kids_filter").(bool)
	if isKids && !movie.IsKidsSafe() {
		return utils.Forbidden(c, "This content is not available on kids profiles")
	}

	cfg := config.Get()
	posterURL := ""
	if movie.PosterPath != "" {
		posterURL = fmt.Sprintf("http://localhost:%s/api/posters/%s", cfg.ServerPort, filepath.Base(movie.PosterPath))
	}

	return utils.SuccessData(c, fiber.Map{
		"movie":      movie,
		"poster_url": posterURL,
		"stream_url": fmt.Sprintf("http://localhost:%s/api/stream/%d", cfg.ServerPort, movie.ID),
	})
}

// HandleUploadMovie handles manual movie file upload with poster
func HandleUploadMovie(c *fiber.Ctx) error {
	title := c.FormValue("title")
	if title == "" {
		return utils.BadRequest(c, "Movie title is required")
	}

	// Get movie file
	movieFile, err := c.FormFile("movie")
	if err != nil {
		return utils.BadRequest(c, "Movie file is required")
	}

	cfg := config.Get()

	// Generate unique filename
	timestamp := time.Now().Unix()
	ext := filepath.Ext(movieFile.Filename)
	safeTitle := sanitizeFilename(title)
	movieFilename := fmt.Sprintf("%d-%s%s", timestamp, safeTitle, ext)
	moviePath := filepath.Join(cfg.DownloadDir, movieFilename)

	// Save movie file
	if err := c.SaveFile(movieFile, moviePath); err != nil {
		log.Printf("❌ Failed to save movie file: %v", err)
		return utils.InternalError(c, "Failed to save movie file")
	}

	// Handle poster (optional)
	posterPath := ""
	posterFile, err := c.FormFile("poster")
	if err == nil && posterFile != nil {
		posterExt := filepath.Ext(posterFile.Filename)
		posterFilename := fmt.Sprintf("%d-%s-poster%s", timestamp, safeTitle, posterExt)
		posterPath = filepath.Join(cfg.PosterDir, posterFilename)
		if err := c.SaveFile(posterFile, posterPath); err != nil {
			log.Printf("⚠️  Failed to save poster: %v", err)
			// Non-fatal: continue without poster
			posterPath = ""
		}
	}

	// Get file size
	fileInfo, _ := os.Stat(moviePath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	// Create movie record
	profileID := auth.GetProfileID(c)
	movie := models.Movie{
		Title:      title,
		FilePath:   moviePath,
		FileSize:   fileSize,
		PosterPath: posterPath,
		Source:     "upload",
		ProfileID:  profileID,
	}

	if err := config.DB.Create(&movie).Error; err != nil {
		log.Printf("❌ Failed to save movie metadata: %v", err)
		return utils.InternalError(c, "Failed to save movie metadata")
	}

	log.Printf("✅ Movie uploaded: %s (%s)", title, formatBytes(fileSize))

	return utils.Success(c, "Movie uploaded successfully", fiber.Map{
		"id":    movie.ID,
		"title": movie.Title,
	})
}

// HandleStreamMovie streams a movie file with HTTP range support
func HandleStreamMovie(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.BadRequest(c, "Invalid movie ID")
	}

	var movie models.Movie
	if err := config.DB.First(&movie, id).Error; err != nil {
		return utils.NotFound(c, "Movie not found")
	}

	// Kids filter
	isKids, _ := c.Locals("kids_filter").(bool)
	if isKids && !movie.IsKidsSafe() {
		return utils.Forbidden(c, "This content is not available on kids profiles")
	}

	// Open file
	file, err := os.Open(movie.FilePath)
	if err != nil {
		log.Printf("❌ Failed to open movie file: %v", err)
		return utils.NotFound(c, "Movie file not found on disk")
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		return utils.InternalError(c, "Failed to read movie file")
	}

	fileSize := stat.Size()

	// Detect content type
	contentType := "video/mp4"
	ext := strings.ToLower(filepath.Ext(movie.FilePath))
	switch ext {
	case ".mkv":
		contentType = "video/x-matroska"
	case ".avi":
		contentType = "video/x-msvideo"
	case ".webm":
		contentType = "video/webm"
	case ".mov":
		contentType = "video/quicktime"
	}

	// Handle range requests for seeking
	rangeHeader := c.Get("Range")
	if rangeHeader != "" {
		return handleRangeRequest(c, file, fileSize, contentType, rangeHeader)
	}

	// Full file response
	c.Set("Content-Type", contentType)
	c.Set("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Set("Accept-Ranges", "bytes")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(movie.FilePath)))

	return c.SendStream(file, int(fileSize))
}

// handleRangeRequest processes HTTP range requests for video seeking
func handleRangeRequest(c *fiber.Ctx, file *os.File, fileSize int64, contentType string, rangeHeader string) error {
	// Parse range header: "bytes=start-end"
	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeHeader, "-")

	var start, end int64

	if parts[0] != "" {
		start, _ = strconv.ParseInt(parts[0], 10, 64)
	}

	if len(parts) > 1 && parts[1] != "" {
		end, _ = strconv.ParseInt(parts[1], 10, 64)
	} else {
		// Stream 2MB chunks for efficient buffering
		end = start + 2*1024*1024 - 1
		if end >= fileSize {
			end = fileSize - 1
		}
	}

	if start >= fileSize {
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		c.Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		return nil
	}

	// Seek to start position
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return utils.InternalError(c, "Failed to seek in file")
	}

	contentLength := end - start + 1

	c.Status(http.StatusPartialContent)
	c.Set("Content-Type", contentType)
	c.Set("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Set("Accept-Ranges", "bytes")

	// Read and send the range
	buf := make([]byte, contentLength)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return utils.InternalError(c, "Failed to read file")
	}

	return c.Send(buf[:n])
}

// HandleDeleteMovie deletes a movie and its files
func HandleDeleteMovie(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.BadRequest(c, "Invalid movie ID")
	}

	var movie models.Movie
	if err := config.DB.First(&movie, id).Error; err != nil {
		return utils.NotFound(c, "Movie not found")
	}

	// Delete movie file
	if movie.FilePath != "" {
		if err := os.Remove(movie.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("⚠️  Failed to delete movie file: %v", err)
		}
	}

	// Delete poster file
	if movie.PosterPath != "" {
		if err := os.Remove(movie.PosterPath); err != nil && !os.IsNotExist(err) {
			log.Printf("⚠️  Failed to delete poster file: %v", err)
		}
	}

	// Delete DB record
	config.DB.Delete(&movie)

	log.Printf("✅ Movie deleted: %s (ID: %d)", movie.Title, movie.ID)
	return utils.Success(c, "Movie deleted successfully", nil)
}

// HandleGetPoster serves a poster image from local storage
func HandleGetPoster(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return utils.BadRequest(c, "Poster name is required")
	}

	cfg := config.Get()
	posterPath := filepath.Join(cfg.PosterDir, name)

	// Security: prevent directory traversal
	absPath, err := filepath.Abs(posterPath)
	if err != nil || !strings.HasPrefix(absPath, cfg.PosterDir) {
		return utils.BadRequest(c, "Invalid poster path")
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return utils.NotFound(c, "Poster not found")
	}

	// Set cache headers (posters don't change)
	c.Set("Cache-Control", "public, max-age=86400") // 24 hours

	return c.SendFile(absPath)
}

// ── Helpers ─────────────────────────────────────────────────────────

func sanitizeFilename(name string) string {
	// Replace unsafe characters with dashes
	replacer := strings.NewReplacer(
		" ", "-", "/", "-", "\\", "-", ":", "-",
		"*", "", "?", "", "\"", "", "<", "", ">", "", "|", "",
	)
	result := replacer.Replace(strings.ToLower(name))
	// Remove consecutive dashes
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
