package auth

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

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// REGISTER HANDLER
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"name":  user.Name,
		"email": user.Email,
	})
}

// LOGIN HANDLER
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"name":    user.Name,
		"email":   user.Email,
		"token":   token,
	})
}

type UpdateStatusRequest struct {
	Status string `json:"onboarding_status"`
}

func (h *Handler) GetOnboardingStatus(c *gin.Context) {
	// Assuming your middleware sets "userID" in the Gin context
	userID, _ := c.Get("userID")

	status, err := h.service.GetOnboardingStatus(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"onboarding_status": status})
}

func (h *Handler) UpdateOnboardingStatus(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req UpdateStatusRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.service.UpdateOnboardingStatus(c.Request.Context(), userID.(string), req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}
