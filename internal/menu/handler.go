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

// PendingMenus returns all menus pending approval
func (h *AdminHandler) PendingMenus(c *gin.Context) {
	// TODO: Implement to fetch all menus with status = 'OCR_COMPLETED'
	c.JSON(http.StatusOK, gin.H{
		"message": "Pending menus endpoint",
		"pending_menus": []gin.H{},
	})
}

// ApproveMenu approves a menu by ID
func (h *AdminHandler) ApproveMenu(c *gin.Context) {
	menuID := c.Param("id")
	
	// TODO: Implement to update menu status to 'APPROVED'
	c.JSON(http.StatusOK, gin.H{
		"message": "Menu approved",
		"menu_id": menuID,
	})
}
