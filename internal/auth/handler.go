package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// temporary in-memory store (TDD step)
var users = make(map[string]bool)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(c *gin.Context) {
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
		})
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "missing required fields",
		})
		return
	}

	if users[req.Email] {
		c.JSON(http.StatusConflict, gin.H{
			"message": "email already exists",
		})
		return
	}

	// mark email as registered
	users[req.Email] = true

	c.JSON(http.StatusCreated, gin.H{
		"id":    "mock-id",
		"name":  req.Name,
		"email": req.Email,
	})
}
