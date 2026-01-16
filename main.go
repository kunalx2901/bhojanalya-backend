package main

import (
	"log"
	"os"

	"bhojanalya/internal/auth"
	"bhojanalya/internal/db"
	"bhojanalya/internal/middleware"
	"bhojanalya/internal/restaurant"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Validate JWT_SECRET early (fail fast)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in .env")
	}

	log.Println("Environment loaded successfully")

	// Create Gin router
	r := gin.Default()

	// Database connection
	pgDB := db.ConnectPostgres()

	// Auth dependencies
	userRepo := auth.NewPostgresUserRepository(pgDB)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	// Public Auth Routes
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	// Protected test route
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

	// RESTAURANT MODULE 

	restaurantRepo := restaurant.NewPostgresRepository(pgDB)
	restaurantService := restaurant.NewService(restaurantRepo)
	restaurantHandler := restaurant.NewHandler(restaurantService)

	restaurantRoutes := r.Group("/restaurants")
	restaurantRoutes.Use(middleware.AuthMiddleware())
	{
		restaurantRoutes.POST("", restaurantHandler.CreateRestaurant)
		restaurantRoutes.GET("/me", restaurantHandler.ListMyRestaurants)
	}

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	log.Println("Server running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
