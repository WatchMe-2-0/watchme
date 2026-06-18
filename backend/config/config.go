package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// AppConfig holds all application configuration
type AppConfig struct {
	// Server
	ServerPort string `json:"server_port"`

	// Database
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`

	// Storage
	DownloadDir string `json:"download_dir"`
	PosterDir   string `json:"poster_dir"`

	// TMDB
	TMDBApiKey string `json:"tmdb_api_key"`

	// Torrent
	MaxConcurrentDownloads int `json:"max_concurrent_downloads"`

	// Auth
	JWTSecret      string `json:"jwt_secret"`
	SessionExpiry  int    `json:"session_expiry_days"`

	// Features
	EnableStreamWhileDownload bool `json:"enable_stream_while_download"`
}

var (
	Cfg     *AppConfig
	cfgMu   sync.RWMutex
	cfgPath string
)

// DefaultConfig returns sensible defaults
func DefaultConfig() *AppConfig {
	homeDir, _ := os.UserHomeDir()
	return &AppConfig{
		ServerPort:                "8000",
		DBHost:                    "localhost",
		DBPort:                    "5432",
		DBUser:                    "admin",
		DBPassword:                "admin",
		DBName:                    "moviesdb",
		DownloadDir:               filepath.Join(homeDir, "watchme", "movies"),
		PosterDir:                 filepath.Join(homeDir, "watchme", "posters"),
		TMDBApiKey:                "",
		MaxConcurrentDownloads:    3,
		JWTSecret:                 "",
		SessionExpiry:             7,
		EnableStreamWhileDownload: true,
	}
}

// LoadConfig reads config from file or creates default
func LoadConfig(path string) *AppConfig {
	cfgPath = path
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("⚙️  No config file found, using defaults")
			// Create config directory and save defaults
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				log.Printf("⚠️  Failed to create config directory: %v", err)
			}
			_ = saveConfig(cfg, path)
		} else {
			log.Printf("⚠️  Failed to read config: %v, using defaults", err)
		}
		Cfg = cfg
		return cfg
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		log.Printf("⚠️  Failed to parse config: %v, using defaults", err)
		cfg = DefaultConfig()
	}

	Cfg = cfg
	return cfg
}

// SaveConfig persists the current config to disk
func SaveConfig() error {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return saveConfig(Cfg, cfgPath)
}

func saveConfig(cfg *AppConfig, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// UpdateConfig atomically updates config fields
func UpdateConfig(updateFn func(cfg *AppConfig)) error {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	updateFn(Cfg)
	return saveConfig(Cfg, cfgPath)
}

// Get returns a read-safe copy of config
func Get() AppConfig {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return *Cfg
}

// EnsureDirectories creates storage directories if they don't exist
func EnsureDirectories() error {
	cfg := Get()
	dirs := []string{cfg.DownloadDir, cfg.PosterDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
