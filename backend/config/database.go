package config

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDatabase initializes PostgreSQL with connection pooling optimized for performance
func ConnectDatabase(cfg *AppConfig) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		SkipDefaultTransaction: true, // Performance: skip wrapping single creates in transactions
		PrepareStmt:            true, // Performance: cache prepared statements
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Configure connection pool for high concurrency
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("❌ Failed to get database instance:", err)
	}

	sqlDB.SetMaxOpenConns(25)                  // Max simultaneous connections
	sqlDB.SetMaxIdleConns(10)                  // Keep 10 idle connections ready
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Recycle connections every 5 min
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)  // Close idle connections after 2 min

	log.Println("✅ Connected to PostgreSQL with connection pooling")

	DB = db
}

// AutoMigrate runs database migrations for all models
func AutoMigrate(models ...interface{}) {
	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			log.Fatalf("❌ Migration failed for %T: %v", model, err)
		}
	}
	log.Println("✅ Database migrations completed")
}
