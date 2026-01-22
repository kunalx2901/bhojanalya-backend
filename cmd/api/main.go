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
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// --------------------
	// ENV SETUP
	// --------------------
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	log.Println("Environment loaded")

	// --------------------
	// DATABASE
	// --------------------
	pgDB := db.ConnectPostgres()

	// cors setup
	r := gin.Default()

// --------------------
// CORS CONFIG (IMPORTANT)
// --------------------
r.Use(cors.New(cors.Config{
	AllowOrigins: []string{
		"http://localhost:3000", // React (CRA)
		"http://localhost:5173", // Vite
		"http://127.0.0.1:5173",
	},
	AllowMethods: []string{
		"GET",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"OPTIONS",
	},
	AllowHeaders: []string{
		"Origin",
		"Content-Type",
		"Authorization",
	},
	ExposeHeaders: []string{
		"Content-Length",
	},
	AllowCredentials: true,
	MaxAge: 12 * time.Hour,
}))


	// --------------------
	// STORAGE (R2)
	// --------------------
	r2Client, err := storage.NewR2Client(context.Background())
	if err != nil {
		log.Fatal("Failed to init R2 client:", err)
	}

	// --------------------
	// AUTH MODULE
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
	// PROTECTED TEST
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
	// RESTAURANT MODULE
	// --------------------
	restaurantRepo := restaurant.NewPostgresRepository(pgDB)
	restaurantService := restaurant.NewService(restaurantRepo)
	restaurantHandler := restaurant.NewHandler(restaurantService)

	restaurantRoutes := r.Group("/restaurants")
	restaurantRoutes.Use(
		middleware.AuthMiddleware(),
		middleware.RequireRole("RESTAURANT"),
	)
	{
		restaurantRoutes.POST("", restaurantHandler.CreateRestaurant)
		restaurantRoutes.GET("/me", restaurantHandler.ListMyRestaurants)
	}

	// --------------------
	// MENU MODULE
	// --------------------
	menuRepo := menu.NewPostgresRepository(pgDB)
	menuService := menu.NewService(menuRepo, r2Client)
	menuHandler := menu.NewHandler(menuService)
	adminMenuHandler := menu.NewAdminHandler(menuService)

	menuRoutes := r.Group("/menus")
	menuRoutes.Use(middleware.AuthMiddleware())
	{
		menuRoutes.POST("/upload", menuHandler.Upload)
	}

	admin := r.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(),
		middleware.RequireRole("ADMIN"),
	)
	{
		admin.GET("/menus/pending", adminMenuHandler.PendingMenus)
		admin.POST("/menus/:id/approve", adminMenuHandler.ApproveMenu)
	}

	// --------------------
	// OCR WORKER (BACKGROUND)
	// --------------------
	
	llmClient := llm.NewGeminiClient()   // ✅ DECLARED HERE

	menuRepo = menu.NewPostgresRepository(pgDB)
	menuService = menu.NewService(menuRepo, r2Client) // ✅ DECLARED HERE
	ocrRepo := ocr.NewRepository(pgDB)


	ocrService := ocr.NewService(
		ocrRepo,
		r2Client,
		llmClient,
		menuService,
	)



	go func() {
		log.Println("OCR worker started")
		if err := ocrService.Start(); err != nil {
			log.Fatal("OCR worker crashed:", err)
		}
	}()

	// --------------------
	// 🔥 LLM TEST ROUTE (DEV ONLY)
	// --------------------
	r.POST("/api/llm/test-gemini", llm.TestGeminiHandler)

	// --------------------
	// HEALTH CHECK
	// --------------------
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// --------------------
	// START SERVER
	// --------------------
	log.Println("Server running on http://localhost:8000")
	if err := r.Run(":8000"); err != nil {
		log.Fatal("Server failed:", err)
	}
}



