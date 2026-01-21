package menu

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

/*
Fake OCR service used only for tests.
It simulates what the worker would eventually do.
*/
type FakeOCRService struct {
	repo Repository
}

func (f *FakeOCRService) Start() error {
	// no-op
	return nil
}

func (f *FakeOCRService) SimulateOCRCompletion(menuID int) error {
	return f.repo.UpdateStatus(
		context.Background(),
		menuID,
		"OCR_DONE",
		nil,
	)
}

func setupMenuTestRouter(repo Repository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	service := NewService(repo, nil)
	handler := NewHandler(service)

	r.POST("/menus/upload", handler.Upload)
	r.GET("/menus/:id/status", handler.GetStatus)

	return r
}

func TestMenuUpload_InitialStatus(t *testing.T) {
	repo := NewInMemoryRepository()
	router := setupMenuTestRouter(repo)

	body := &bytes.Buffer{}
	req, _ := http.NewRequest("POST", "/menus/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	menuID := int(resp["id"].(float64))

	menu, err := repo.GetMenuUpload(context.Background(), menuID)
	if err != nil {
		t.Fatal(err)
	}

	if menu.Status != "MENU_UPLOADED" {
		t.Fatalf("expected MENU_UPLOADED, got %s", menu.Status)
	}
}
