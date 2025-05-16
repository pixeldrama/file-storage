package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/benjamin/file-storage-go/pkg/config"
	"github.com/benjamin/file-storage-go/pkg/database"
)

func main() {

	os.Setenv("SKIP_STORAGE_VALIDATION", "true")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	migrationsPath, err := filepath.Abs("migrations")
	if err != nil {
		log.Fatalf("Failed to get migrations path: %v", err)
	}

	if err := database.RunMigrations(cfg.GetDBConnString(), migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed successfully")
}
