package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectPostgres() *pgxpool.Pool {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal(err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(context.Background()); err != nil {
		log.Fatal("Postgres connection failed:", err)
	}

	log.Println("✅ Connected to Aiven PostgreSQL")

	// Initialize schema
	if err := initSchema(db); err != nil {
		log.Fatal("Failed to initialize schema:", err)
	}

	return db
}

// initSchema creates or updates the database schema
func initSchema(db *pgxpool.Pool) error {
	ctx := context.Background()

	// -------------------------------
	// USERS
	// -------------------------------
	userTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'RESTAURANT',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(ctx, userTableSQL); err != nil {
		return err
	}

	addRoleColumnSQL := `
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'RESTAURANT'
	`
	if _, err := db.Exec(ctx, addRoleColumnSQL); err != nil {
		log.Println("Note: role column may already exist")
	}

	// -------------------------------
	// MENU UPLOADS
	// -------------------------------
	menuUploadsSQL := `
		CREATE TABLE IF NOT EXISTS menu_uploads (
			id SERIAL PRIMARY KEY,
			restaurant_id UUID NOT NULL,
			image_url VARCHAR(500) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'MENU_UPLOADED',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (restaurant_id) REFERENCES users(id)
		)
	`
	if _, err := db.Exec(ctx, menuUploadsSQL); err != nil {
		return err
	}

	// -------------------------------
	// ADMIN APPROVAL COLUMNS (FINAL PHASE)
	// -------------------------------
	approvalColumnsSQL := `
		ALTER TABLE menu_uploads
		ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP NULL;

		ALTER TABLE menu_uploads
		ADD COLUMN IF NOT EXISTS approved_by UUID NULL;

		ALTER TABLE menu_uploads
		ADD COLUMN IF NOT EXISTS rejection_reason TEXT NULL;
	`
	if _, err := db.Exec(ctx, approvalColumnsSQL); err != nil {
		return err
	}

	log.Println("✅ Schema initialized successfully")
	return nil
}
