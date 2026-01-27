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
			"help":  "Include 'restaurant_id' field in your form-data",
		})
		return
	}

	// Convert to int
	var restaurantID int
	if _, err := fmt.Sscanf(restaurantIDStr, "%d", &restaurantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid restaurant_id format. Must be a number",
			"details": err.Error(),
		})
		return
	}

	if restaurantID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "restaurant_id must be a positive number",
		})
		return
	}

	// Optional: Verify user owns this restaurant (security check)
	if userID, exists := c.Get("user_id"); exists {
		fmt.Printf("[UPLOAD] User %s uploading to restaurant %d\n", userID, restaurantID)
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

	menuID, objectKey, err := h.service.UploadMenu(
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
		"object_key":     objectKey,
		"restaurant_id":  restaurantID,
		"status":         "MENU_UPLOADED",
		"message":        "Menu uploaded. OCR and parsing will start automatically.",
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

	var menuID int
	if _, err := fmt.Sscanf(menuIDStr, "%d", &menuID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid menu id",
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
		menuID,
		adminID,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "approved",
		"menu_id": menuID,
	})
}
