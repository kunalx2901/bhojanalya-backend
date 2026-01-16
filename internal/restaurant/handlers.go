package restaurant

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

// handler for creating a restaurant
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

	// 🔐 Extract user ID from JWT context
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


// get restaurants for the logged-in owner
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
