package menu

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// handler to post a new menu for a restaurant
func (h *Handler) UploadMenu(c *gin.Context) {
	restaurantID := c.Param("id")

	file, err := c.FormFile("menu")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "menu file required"})
		return
	}

	// 1️⃣ Validate extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExt := map[string]bool{
		".pdf":  true,
		".txt":  true,
		".csv":  true,
		".json": true,
		".xml":  true,
	}
	if !allowedExt[ext] {
		c.JSON(400, gin.H{"error": "unsupported file type"})
		return
	}

	// 2️⃣ Validate file size
	const maxSize = 5 << 20 // 5MB
	if file.Size > maxSize {
		c.JSON(400, gin.H{"error": "file exceeds 5MB limit"})
		return
	}

	// 3️⃣ Validate MIME type
	src, err := file.Open()
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid file"})
		return
	}
	defer src.Close()

	buf := make([]byte, 512)
	if _, err := src.Read(buf); err != nil {
		c.JSON(400, gin.H{"error": "invalid file"})
		return
	}

	mime := http.DetectContentType(buf)
	allowedMIME := map[string]bool{
		"application/pdf":  true,
		"text/plain":       true,
		"text/csv":         true,
		"application/json": true,
		"application/xml":  true,
	}
	if !allowedMIME[mime] {
		c.JSON(400, gin.H{"error": "file type not supported"})
		return
	}

	// 4️⃣ Block media & archives
	if strings.HasPrefix(mime, "audio/") ||
		strings.HasPrefix(mime, "video/") ||
		mime == "application/zip" ||
		mime == "application/octet-stream" {
		c.JSON(400, gin.H{"error": "media and archive files are not allowed"})
		return
	}

	// 5️⃣ Prepare storage
	if err := os.MkdirAll("uploads/menus", os.ModePerm); err != nil {
		c.JSON(500, gin.H{"error": "failed to prepare storage"})
		return
	}

	fileName := uuid.New().String() + ext
	filePath := filepath.Join("uploads/menus", fileName)

	// 6️⃣ Save file FIRST
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(500, gin.H{"error": "failed to save file"})
		return
	}

	// 7️⃣ Save DB record AFTER file is saved
	if err := h.service.UploadMenu(restaurantID, filePath); err != nil {
		_ = os.Remove(filePath) // rollback
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "menu uploaded"})
}
