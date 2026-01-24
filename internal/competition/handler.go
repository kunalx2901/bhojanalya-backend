package competition

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

// POST /admin/competition/recompute
func (h *Handler) Recompute(c *gin.Context) {
	var req struct {
		City        string `json:"city"`
		CuisineType string `json:"cuisine_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if req.City == "" || req.CuisineType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city and cuisine_type required"})
		return
	}

	if err := h.service.RecomputeSnapshot(
		c.Request.Context(),
		req.City,
		req.CuisineType,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// GET /competition/insights
func (h *Handler) Get(c *gin.Context) {
	city := c.Query("city")
	cuisine := c.Query("cuisine_type")

	if city == "" || cuisine == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "city and cuisine_type required",
		})
		return
	}

	snapshot, err := h.service.GetSnapshot(
		c.Request.Context(),
		city,
		cuisine,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no data available",
		})
		return
	}

	c.JSON(http.StatusOK, snapshot)
}
