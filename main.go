package main

import (
	"log"
	"os"

	"bhojanalya/internal/auth"
	"bhojanalya/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	

)

func main() {
	// 1️⃣ Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ Error loading .env file")
	}

	// 2️⃣ Validate JWT_SECRET early (fail fast)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET is not set in .env")
	}

	log.Println("✅ Environment loaded successfully")

	// 3️⃣ Create Gin router
	r := gin.Default()

	// 4️⃣ Setup Auth dependencies
	repo := auth.NewInMemoryUserRepository()
	service := auth.NewService(repo)
	handler := auth.NewHandler(service)

	// 5️⃣ Public Auth Routes
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", handler.Register)
		authRoutes.POST("/login", handler.Login)
	}

	// 6️⃣ Protected Routes (JWT required)
	protected := r.Group("/protected")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/ping", func(c *gin.Context) {
			email, _ := c.Get("userEmail")
			c.JSON(200, gin.H{
				"message": "authenticated",
				"email":   email,
			})
		})
	}

	// 7️⃣ Health Check (always useful)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 8️⃣ Start server
	log.Println("Server running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
