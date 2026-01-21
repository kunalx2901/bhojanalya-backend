package menu

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Upload(c *gin.Context) {
	restaurantID := c.GetInt("userID")

	file, header, err := c.Request.FormFile("menu_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "menu_file is required"})
		return
	}
	defer file.Close()

	// 🔐 VALIDATE EXTENSION HERE
	if err := ValidateFileExtension(header.Filename); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	menuID, url, err := h.service.UploadMenu(
		c.Request.Context(),
		restaurantID,
		file,
		header.Filename,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"menu_upload_id": menuID,
		"object_key": url,
		"status": "MENU_UPLOADED",
		"message": "Menu uploaded. OCR processing will start automatically.",
	})
}
