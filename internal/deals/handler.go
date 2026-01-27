package deals

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

//
// --------------------------------------------------
// GET /restaurants/:id/deals/suggestions
// --------------------------------------------------
//

func (h *Handler) GetDealSuggestion() gin.HandlerFunc {
	return func(c *gin.Context) {

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

		suggestion, err := h.service.GetDealSuggestion(
			c.Request.Context(),
			restaurantID,
			userID.(string),
		)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, suggestion)
	}
}

//
// --------------------------------------------------
// POST /restaurants/:id/deals
// --------------------------------------------------
//

func (h *Handler) CreateDeal() gin.HandlerFunc {
	return func(c *gin.Context) {

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

		var deal Deal
		if err := c.ShouldBindJSON(&deal); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// 🔒 Enforce restaurant ownership via path
		deal.RestaurantID = restaurantID

		if err := h.service.CreateDeal(
			c.Request.Context(),
			userID.(string),
			&deal,
		); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "deal created and pending approval",
			"deal":    deal,
		})
	}
}
