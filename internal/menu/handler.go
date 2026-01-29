package menu

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

type AdminHandler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func NewAdminHandler(service *Service) *AdminHandler {
	return &AdminHandler{service: service}
}

// --------------------------------------------------
// Restaurant uploads menu
// --------------------------------------------------
func (h *Handler) Upload(c *gin.Context) {
	// Get restaurant_id from form data
	restaurantIDStr := c.PostForm("restaurant_id")
	if restaurantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "restaurant_id is required in form data",
		})
		return
	}

	var restaurantID int
	if _, err := fmt.Sscanf(restaurantIDStr, "%d", &restaurantID); err != nil || restaurantID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid restaurant_id",
		})
		return
	}

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

	objectKey, err := h.service.UploadMenu(
		c.Request.Context(),
		restaurantID,
		file,
		header.Filename,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"restaurant_id": restaurantID,
		"object_key":    objectKey,
		"status":        "MENU_UPLOADED",
		"message":       "Menu uploaded successfully. OCR and parsing will start automatically.",
	})
}

// --------------------------------------------------
// Admin: view parsed menus pending approval
// --------------------------------------------------
func (h *AdminHandler) PendingMenus(c *gin.Context) {
	menus, err := h.service.GetPendingMenus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, menus)
}

// --------------------------------------------------
// Admin: approve menu
// --------------------------------------------------
func (h *AdminHandler) ApproveMenu(c *gin.Context) {
	menuIDStr := c.Param("id")

	var restaurantID int
	if _, err := fmt.Sscanf(menuIDStr, "%d", &restaurantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid restaurant id",
		})
		return
	}

	adminID := c.GetString("userID")
	if adminID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "admin user not found in context",
		})
		return
	}

	if err := h.service.ApproveMenu(
		c.Request.Context(),
		restaurantID,
		adminID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "approved",
		"restaurant_id": restaurantID,
	})
}
