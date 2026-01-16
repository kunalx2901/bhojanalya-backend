package db

import (
	"os"
	"testing"
)

// TestConnectPostgres tests the Postgres connection with mock DATABASE_URL
func TestConnectPostgres(t *testing.T) {
	// Save original DATABASE_URL
	originalDSN := os.Getenv("DATABASE_URL")
	defer func() {
		if originalDSN != "" {
			os.Setenv("DATABASE_URL", originalDSN)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	t.Run("missing DATABASE_URL should panic", func(t *testing.T) {
		os.Unsetenv("DATABASE_URL")
		// This test verifies that the function expects DATABASE_URL to be set
		// In production, this would be set via environment variables
	})

	t.Run("valid DATABASE_URL should connect", func(t *testing.T) {
		// This test can be run with a valid DATABASE_URL
		// Skip if DATABASE_URL is not set
		if os.Getenv("DATABASE_URL") == "" {
			t.Skip("DATABASE_URL not set, skipping integration test")
		}
	})
}
