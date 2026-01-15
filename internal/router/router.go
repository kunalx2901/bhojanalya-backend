package router

import (
	"github.com/gin-gonic/gin"
	"bhojanalya/internal/auth"
)

// NewRouter creates and returns the main router
func NewRouter() *gin.Engine {
	r := gin.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", auth.Register)
	}

	return r
}
