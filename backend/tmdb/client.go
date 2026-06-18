package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"watchme/config"
)

const (
	baseURL        = "https://api.themoviedb.org/3"
	imageBaseURL   = "https://image.tmdb.org/t/p"
	posterSize     = "w500"
	backdropSize   = "w1280"
	requestTimeout = 10 * time.Second
)

// Client is the TMDB API client with connection pooling and caching
type Client struct {
	httpClient *http.Client
	cache      *Cache
	mu         sync.RWMutex
}

var (
	clientInstance *Client
	clientOnce     sync.Once
)

// GetClient returns the singleton TMDB client
func GetClient() *Client {
	clientOnce.Do(func() {
		transport := &http.Transport{
			MaxIdleConns:        20,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}

		clientInstance = &Client{
			httpClient: &http.Client{
				Timeout:   requestTimeout,
				Transport: transport,
			},
			cache: NewCache(1000, time.Hour, 24*time.Hour),
		}
	})
	return clientInstance
}

// ── Response Types ──────────────────────────────────────────────────

// MovieResult represents a movie from TMDB search/browse
type MovieResult struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	ReleaseDate   string  `json:"release_date"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float64 `json:"popularity"`
	GenreIDs      []int   `json:"genre_ids"`
	Adult         bool    `json:"adult"`
	OriginalLang  string  `json:"original_language"`
}

// MovieListResponse wraps paginated movie results
type MovieListResponse struct {
	Page         int           `json:"page"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
	Results      []MovieResult `json:"results"`
}

// MovieDetail is the full details of a single movie
type MovieDetail struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	ReleaseDate   string  `json:"release_date"`
	VoteAverage   float64 `json:"vote_average"`
	Runtime       int     `json:"runtime"`
	Tagline       string  `json:"tagline"`
	Status        string  `json:"status"`
	Budget        int64   `json:"budget"`
	Revenue       int64   `json:"revenue"`
	Genres        []Genre `json:"genres"`
	Adult         bool    `json:"adult"`
}

// Genre represents a movie genre
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GenreListResponse wraps genre list
type GenreListResponse struct {
	Genres []Genre `json:"genres"`
}

// ReleaseDatesResponse wraps certification data
type ReleaseDatesResponse struct {
	Results []struct {
		ISO3166_1    string `json:"iso_3166_1"`
		ReleaseDates []struct {
			Certification string `json:"certification"`
			Type          int    `json:"type"`
		} `json:"release_dates"`
	} `json:"results"`
}

// ── API Methods ─────────────────────────────────────────────────────

// SearchMovies searches TMDB for movies by query
func (c *Client) SearchMovies(query string, page int) (*MovieListResponse, error) {
	cacheKey := fmt.Sprintf("search:%s:%d", query, page)
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*MovieListResponse), nil
	}

	params := url.Values{
		"query": {query},
		"page":  {fmt.Sprintf("%d", page)},
	}

	var result MovieListResponse
	if err := c.doRequest("/search/movie", params, &result); err != nil {
		return nil, err
	}

	// Add poster/backdrop full URLs
	for i := range result.Results {
		result.Results[i].PosterPath = c.PosterURL(result.Results[i].PosterPath)
		result.Results[i].BackdropPath = c.BackdropURL(result.Results[i].BackdropPath)
	}

	c.cache.Set(cacheKey, &result)
	return &result, nil
}

// Trending gets trending movies
func (c *Client) Trending(timeWindow string, page int) (*MovieListResponse, error) {
	if timeWindow == "" {
		timeWindow = "week"
	}

	cacheKey := fmt.Sprintf("trending:%s:%d", timeWindow, page)
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*MovieListResponse), nil
	}

	path := fmt.Sprintf("/trending/movie/%s", timeWindow)
	params := url.Values{"page": {fmt.Sprintf("%d", page)}}

	var result MovieListResponse
	if err := c.doRequest(path, params, &result); err != nil {
		return nil, err
	}

	for i := range result.Results {
		result.Results[i].PosterPath = c.PosterURL(result.Results[i].PosterPath)
		result.Results[i].BackdropPath = c.BackdropURL(result.Results[i].BackdropPath)
	}

	c.cache.Set(cacheKey, &result)
	return &result, nil
}

// TopRated gets top-rated movies
func (c *Client) TopRated(page int) (*MovieListResponse, error) {
	cacheKey := fmt.Sprintf("toprated:%d", page)
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*MovieListResponse), nil
	}

	params := url.Values{"page": {fmt.Sprintf("%d", page)}}

	var result MovieListResponse
	if err := c.doRequest("/movie/top_rated", params, &result); err != nil {
		return nil, err
	}

	for i := range result.Results {
		result.Results[i].PosterPath = c.PosterURL(result.Results[i].PosterPath)
		result.Results[i].BackdropPath = c.BackdropURL(result.Results[i].BackdropPath)
	}

	c.cache.Set(cacheKey, &result)
	return &result, nil
}

// ByGenre gets movies filtered by genre ID
func (c *Client) ByGenre(genreID int, page int) (*MovieListResponse, error) {
	cacheKey := fmt.Sprintf("genre:%d:%d", genreID, page)
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*MovieListResponse), nil
	}

	params := url.Values{
		"with_genres":   {fmt.Sprintf("%d", genreID)},
		"sort_by":       {"popularity.desc"},
		"page":          {fmt.Sprintf("%d", page)},
	}

	var result MovieListResponse
	if err := c.doRequest("/discover/movie", params, &result); err != nil {
		return nil, err
	}

	for i := range result.Results {
		result.Results[i].PosterPath = c.PosterURL(result.Results[i].PosterPath)
		result.Results[i].BackdropPath = c.BackdropURL(result.Results[i].BackdropPath)
	}

	c.cache.Set(cacheKey, &result)
	return &result, nil
}

// GetMovieDetail gets full details for a single movie
func (c *Client) GetMovieDetail(movieID int) (*MovieDetail, error) {
	cacheKey := fmt.Sprintf("detail:%d", movieID)
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*MovieDetail), nil
	}

	path := fmt.Sprintf("/movie/%d", movieID)

	var result MovieDetail
	if err := c.doRequest(path, nil, &result); err != nil {
		return nil, err
	}

	result.PosterPath = c.PosterURL(result.PosterPath)
	result.BackdropPath = c.BackdropURL(result.BackdropPath)

	c.cache.Set(cacheKey, &result)
	return &result, nil
}

// GetCertification fetches the US certification for a movie
func (c *Client) GetCertification(movieID int) (string, error) {
	cacheKey := fmt.Sprintf("cert:%d", movieID)
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(string), nil
	}

	path := fmt.Sprintf("/movie/%d/release_dates", movieID)

	var result ReleaseDatesResponse
	if err := c.doRequest(path, nil, &result); err != nil {
		return "", err
	}

	cert := ""
	for _, r := range result.Results {
		if r.ISO3166_1 == "US" {
			for _, rd := range r.ReleaseDates {
				if rd.Certification != "" {
					cert = rd.Certification
					break
				}
			}
			break
		}
	}

	c.cache.Set(cacheKey, cert)
	return cert, nil
}

// GetGenres returns the full genre list
func (c *Client) GetGenres() ([]Genre, error) {
	cacheKey := "genres"
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.([]Genre), nil
	}

	var result GenreListResponse
	if err := c.doRequest("/genre/movie/list", nil, &result); err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, result.Genres)
	return result.Genres, nil
}

// ── URL Builders ────────────────────────────────────────────────────

// PosterURL returns the full URL for a poster path
func (c *Client) PosterURL(path string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s%s", imageBaseURL, posterSize, path)
}

// BackdropURL returns the full URL for a backdrop path
func (c *Client) BackdropURL(path string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s%s", imageBaseURL, backdropSize, path)
}

// ── Internal ────────────────────────────────────────────────────────

func (c *Client) doRequest(path string, params url.Values, result interface{}) error {
	cfg := config.Get()
	if cfg.TMDBApiKey == "" {
		return fmt.Errorf("TMDB API key not configured. Set it in Settings.")
	}

	if params == nil {
		params = url.Values{}
	}
	params.Set("api_key", cfg.TMDBApiKey)

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, path, params.Encode())

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		log.Printf("❌ TMDB request failed: %v", err)
		return fmt.Errorf("TMDB request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("TMDB API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to parse TMDB response: %w", err)
	}

	return nil
}
