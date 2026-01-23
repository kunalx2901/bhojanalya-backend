package menu

import (
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
	restaurantID := c.GetInt("userID")

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
