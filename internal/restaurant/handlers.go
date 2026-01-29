package restaurant

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// --------------------------------------------------
// Create restaurant
// --------------------------------------------------
func (h *Handler) CreateRestaurant(c *gin.Context) {
	var req struct {
		Name             string `json:"name"`
		City             string `json:"city"`
		CuisineType      string `json:"cuisine_type"`
		ShortDescription string `json:"short_description"`
		OpensAt          string `json:"opens_at"`
		ClosesAt         string `json:"closes_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	restaurant, err := h.service.CreateRestaurant(
		req.Name,
		req.City,
		req.CuisineType,
		req.ShortDescription,
		req.OpensAt,
		req.ClosesAt,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, restaurant)
}

// --------------------------------------------------
// List restaurants owned by user
// --------------------------------------------------
func (h *Handler) ListMyRestaurants(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	restaurants, err := h.service.ListMyRestaurants(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch restaurants"})
		return
	}

	c.JSON(http.StatusOK, restaurants)
}

// --------------------------------------------------
// ADMIN: List approved restaurants
// --------------------------------------------------
func (h *Handler) ListApprovedRestaurants(c *gin.Context) {
	restaurants, err := h.service.ListApprovedRestaurants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch approved restaurants"})
		return
	}

	c.JSON(http.StatusOK, restaurants)
}

// --------------------------------------------------
// ADMIN: View restaurant details
// --------------------------------------------------
func (h *Handler) GetAdminRestaurantDetails(c *gin.Context) {
	var restaurantID int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &restaurantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid restaurant id"})
		return
	}

	details, err := h.service.GetAdminRestaurantDetails(
		c.Request.Context(),
		restaurantID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, details)
}

// --------------------------------------------------
// Get competitive insight
// --------------------------------------------------
func (h *Handler) GetCompetitionInsight(c *gin.Context) {
	var restaurantID int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &restaurantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid restaurant id"})
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	insight, err := h.service.GetCompetitiveInsight(
		c.Request.Context(),
		restaurantID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, insight)
}

// --------------------------------------------------
// POST /restaurants/:id/images
// --------------------------------------------------
func (h *Handler) UploadImages(c *gin.Context) {
	var restaurantID int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &restaurantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid restaurant id"})
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil || form.File["images"] == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "images are required"})
		return
	}

	if err := h.service.UploadImages(
		c.Request.Context(),
		restaurantID,
		userID,
		form.File["images"],
	); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "images uploaded successfully",
	})
}

// --------------------------------------------------
// GET /restaurants/:id/preview
// --------------------------------------------------
func (h *Handler) Preview(c *gin.Context) {
	var restaurantID int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &restaurantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid restaurant id"})
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return
	}

	data, err := h.service.GetPreview(
		c.Request.Context(),
		restaurantID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}


// to get the restaurant approved by the admin 
func (h *Handler) ApproveRestaurant(c *gin.Context) {
	var restaurantID int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &restaurantID); err != nil {
		c.JSON(400, gin.H{"error": "invalid restaurant id"})
		return
	}

	adminID := c.GetString("userID")
	if adminID == "" {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.service.ApproveRestaurant(
		c.Request.Context(),
		restaurantID,
		adminID,
	); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":       "restaurant, menu, and deals approved",
		"restaurant_id": restaurantID,
	})
}
