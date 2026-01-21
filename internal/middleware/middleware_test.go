package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bhojanalya/internal/auth"

	"github.com/gin-gonic/gin"
)

// TestAuthMiddleware_MissingAuthHeader tests the middleware with missing Authorization header
func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestAuthMiddleware_InvalidAuthFormat tests the middleware with invalid Bearer format
func TestAuthMiddleware_InvalidAuthFormat(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestAuthMiddleware_InvalidToken tests the middleware with an invalid token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_xyz")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestAuthMiddleware_ValidToken tests the middleware with a valid token
func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Set JWT_SECRET for testing
	t.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")

	// Generate a valid test token
	token, err := auth.GenerateToken("test-user-id", "test@example.com", "restaurant")
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		userEmail, _ := c.Get("userEmail")
		c.JSON(http.StatusOK, gin.H{
			"userID":    userID,
			"userEmail": userEmail,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
