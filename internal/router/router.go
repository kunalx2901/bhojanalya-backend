package router

import (
	"bhojanalya/internal/auth"

	"github.com/gin-gonic/gin"
)

func NewRouter(service *auth.Service) *gin.Engine {
	r := gin.Default()

	// Health check route
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/auth/register", func(c *gin.Context) {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		user, err := service.Register(req.Name, req.Email, req.Password)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, gin.H{
			"name":  user.Name,
			"email": user.Email,
		})
	})

	return r
}
