package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/benjamin/file-storage-go/pkg/config"
	"github.com/benjamin/file-storage-go/pkg/database"
)

func main() {
	// Set environment variable to skip storage validation
	os.Setenv("SKIP_STORAGE_VALIDATION", "true")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get the absolute path to the migrations directory
	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		log.Fatalf("Failed to get migrations path: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(cfg.GetDBConnString(), migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed successfully")
}
