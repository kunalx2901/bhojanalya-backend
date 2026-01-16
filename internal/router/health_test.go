package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bhojanalya/internal/auth"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	
	// Create mock service for testing
	repo := auth.NewInMemoryUserRepository()
	service := auth.NewService(repo)
	r := NewRouter(service)

	// ❌ No route registered yet (RED phase)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Act
	r.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
