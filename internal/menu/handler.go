package menu

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Upload(c *gin.Context) {
	// Note: You might want to ensure you're getting the correct ID type 
	// based on how your AuthMiddleware sets it.
	restaurantID := c.GetInt("restaurantID") 

	file, header, err := c.Request.FormFile("menu_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "menu_file is required"})
		return
	}
	defer file.Close()

	if err := ValidateFileExtension(header.Filename); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		"object_key":     url,
		"status":         StatusMenuUploaded, // Using the constant from model.go
		"message":        "Menu uploaded. OCR and Price Analysis will start automatically.",
	})
}

// GetStatus allows the frontend to poll for the structured JSON result
func (h *Handler) GetStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu id"})
		return
	}

	// You will need to implement GetMenuUpload in your service/repository
	menuUpload, err := h.service.GetMenuUpload(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu upload not found"})
		return
	}

	// This response will now include the "structured_data" JSON from Hugging Face 
	// once the status is "OCR_DONE"
	c.JSON(http.StatusOK, menuUpload)
}