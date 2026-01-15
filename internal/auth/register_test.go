package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/auth/register", Register)

	return r
}
func TestRegisterSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupTestRouter()

	payload := map[string]string{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "Password@123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

func TestRegisterMissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupTestRouter()

	payload := map[string]string{
		"email": "test@example.com",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupTestRouter()

	payload := map[string]string{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "Password@123",
	}

	body, _ := json.Marshal(payload)

	// First request (should succeed)
	req1 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// Second request (should fail)
	req2 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w2.Code)
	}
}
