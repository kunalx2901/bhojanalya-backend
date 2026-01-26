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
		Name        string `json:"name"`
		City        string `json:"city"`
		CuisineType string `json:"cuisine_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ownerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	restaurant, err := h.service.CreateRestaurant(
		req.Name,
		req.City,
		req.CuisineType,
		ownerID.(string),
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
	ownerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	restaurants, err := h.service.ListMyRestaurants(ownerID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch restaurants"})
		return
	}

	c.JSON(http.StatusOK, restaurants)
}

// --------------------------------------------------
// Get competitive insight for restaurant
// --------------------------------------------------
func (h *Handler) GetCompetitionInsight(c *gin.Context) {
	var restaurantID int
	if _, err := fmt.Sscanf(c.Param("id"), "%d", &restaurantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid restaurant id"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	insight, err := h.service.GetCompetitiveInsight(
		c.Request.Context(),
		restaurantID,
		userID.(string),
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, insight)
}

