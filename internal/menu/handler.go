package menu

import (
	"fmt"
	"net/http"
	"strconv"

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
			"error": "Invalid restaurant_id format. Must be a number",
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
	userID, exists := c.Get("userID")
	if exists {
		// You could add a check here to verify the user owns this restaurant
		// For now, we'll just log it for debugging
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
	// TODO (NEXT PHASE):
	// Fetch menus where status = 'PARSED' AND approved_at IS NULL
	c.JSON(http.StatusOK, gin.H{
		"pending_menus": []gin.H{},
	})
}

// --------------------------------------------------
// Admin: approve menu
// --------------------------------------------------
func (h *AdminHandler) ApproveMenu(c *gin.Context) {
	menuID := c.Param("id")

	// TODO (NEXT PHASE):
	// Mark menu as approved and unlock deal suggestions
	c.JSON(http.StatusOK, gin.H{
		"message": "Menu approved",
		"menu_id": menuID,
	})
}