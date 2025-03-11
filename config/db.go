package config

import (
	"fmt"
	"log"

	"backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

	// Load environment variables (or use hardcoded values for now)
	dsn := fmt.Sprintf(
		"host=localhost user=admin password=admin dbname=moviesdb port=5432 sslmode=disable",
	)

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	log.Println("✅ Connected to PostgreSQL successfully")

	// Auto-migrate the models
	err = db.AutoMigrate(&models.Movie{})
	if err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	log.Println("✅ Database migration completed")

	// Assign to global DB variable
	DB = db
}
