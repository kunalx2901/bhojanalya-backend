package main

import (
	"fmt"
	"os"
	"testing"

	"bhojanalya/internal/auth"

	"github.com/google/uuid"
)

func TestJWTFlow(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test-secret-key-12345")

	userID := uuid.New().String()
	email := "test@example.com"

	// Generate token
	token, err := auth.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	fmt.Printf("Generated token: %s\n", token)

	// Validate token
	extractedUserID, extractedEmail, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	fmt.Printf("Extracted userID: %s\n", extractedUserID)
	fmt.Printf("Extracted email: %s\n", extractedEmail)

	if extractedUserID != userID {
		t.Fatalf("Expected userID %s, got %s", userID, extractedUserID)
	}

	if extractedEmail != email {
		t.Fatalf("Expected email %s, got %s", email, extractedEmail)
	}

	fmt.Println("✅ JWT flow works correctly!")
}
