package main

import (
	"bhojanalya/internal/auth"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	repo := auth.NewInMemoryUserRepository()
	service := auth.NewService(repo)
	handler := auth.NewHandler(service)

	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", handler.Register)
		authRoutes.POST("/login", handler.Login)
	}

	r.Run(":8080")
}
