package main

import (
	"bhojanalya/internal/auth"
	"bhojanalya/internal/db"
	"bhojanalya/internal/llm"
	"bhojanalya/internal/menu"
	"bhojanalya/internal/middleware"
	"bhojanalya/internal/ocr"
	"bhojanalya/internal/restaurant"
	"bhojanalya/internal/storage"
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables (only in dev)
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, relying on environment variables")
		}
	}

	// Validate JWT_SECRET early
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	log.Println("Environment loaded successfully")

	// --------------------
	// Initialize R2 Client (needed by menu and OCR modules)
	// --------------------
	r2Client, err := storage.NewR2Client(context.Background())
	if err != nil {
		log.Fatal("Failed to initialize R2 client:", err)
	}

	// --------------------
	// Database
	// --------------------
	pgDB := db.ConnectPostgres()

	// --------------------
	// Gin setup
	// --------------------
	r := gin.Default()

	// --------------------
	// Auth module
	// --------------------
	userRepo := auth.NewPostgresUserRepository(pgDB)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	// --------------------
	// Protected test route
	// --------------------
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

	// --------------------
	// Restaurant module
	// --------------------
	restaurantRepo := restaurant.NewPostgresRepository(pgDB)
	restaurantService := restaurant.NewService(restaurantRepo)
	restaurantHandler := restaurant.NewHandler(restaurantService)

	restaurantRoutes := r.Group("/restaurants")
	restaurantRoutes.Use(middleware.AuthMiddleware())
	{
		restaurantRoutes.POST("", restaurantHandler.CreateRestaurant)
		restaurantRoutes.GET("/me", restaurantHandler.ListMyRestaurants)
	}

	// --------------------
	// Menu module
	// --------------------
	menuRepo := menu.NewPostgresRepository(pgDB)
	menuService := menu.NewService(menuRepo, r2Client)
	menuHandler := menu.NewHandler(menuService)

	menus := r.Group("/menus")
	menus.Use(middleware.AuthMiddleware())
	{
		menus.POST("/upload", menuHandler.Upload)
	}

	// --------------------
	// OCR WORKER (CRITICAL)
	// --------------------
	ocrRepo := ocr.NewRepository(pgDB)
	ocrService := ocr.NewService(ocrRepo, r2Client)

	// ✅ MUST run in goroutine
	go func() {
		log.Println("OCR worker started")
		if err := ocrService.Start(); err != nil {
			log.Fatal("OCR worker crashed:", err)
		}
	}()

	// test file for LLM module
	llamaClient := llm.NewLLaMAClient()
	r.GET("/test/llama", llm.TestLLaMA(llamaClient))

	// --------------------
	// Health
	// --------------------
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// --------------------
	// Start server
	// --------------------
	log.Println("Server running on http://localhost:8000")
	if err := r.Run(":8000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
