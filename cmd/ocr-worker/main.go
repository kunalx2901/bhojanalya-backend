package main

import (
	"context"
	"log"
	"os"
	"time"

	"bhojanalya/internal/db"
	"bhojanalya/internal/ocr"
	"bhojanalya/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, using environment variables")
	}

	log.Println("🧠 OCR Worker starting...")

	// Validate database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in .env")
	}

	// Database connection
	pgDB := db.ConnectPostgres()

	log.Println("✅ Connected to PostgreSQL")

	// Initialize R2 storage client
	ctx := context.Background()
	r2Client, err := storage.NewR2Client(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize R2 client: %v", err)
	}

	// Initialize OCR service
	repo := ocr.NewRepository(pgDB)
	service := ocr.NewService(repo, r2Client)

	log.Println("✅ OCR Worker initialized and running...")
	log.Println("Processing menu uploads every 2 seconds. Press Ctrl+C to stop.")

	// Process menu uploads indefinitely
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := service.ProcessOne()
		if err != nil {
			log.Printf("⚠️  OCR error: %v", err)
		}
	}
}
