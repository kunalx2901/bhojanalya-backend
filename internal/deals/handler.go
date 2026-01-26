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

// --------------------------------------------------
// GET /restaurants/:id/deals
// --------------------------------------------------
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
