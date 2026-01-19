package menu

import (
	"net/http"
	"os"
	"path/filepath"

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

	// Ensure upload directory exists
	if err := os.MkdirAll("uploads/menus", os.ModePerm); err != nil {
		c.JSON(500, gin.H{"error": "failed to prepare storage"})
		return
	}

	fileName := uuid.New().String() + filepath.Ext(file.Filename)
	filePath := filepath.Join("uploads/menus", fileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(500, gin.H{"error": "failed to save file"})
		return
	}

	if err := h.service.UploadMenu(restaurantID, filePath); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "menu uploaded"})
}
