package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"bhojanalya/internal/auth"
	"bhojanalya/internal/db"
	"bhojanalya/internal/llm"
	"bhojanalya/internal/menu"
	"bhojanalya/internal/middleware"
	"bhojanalya/internal/ocr"
	"bhojanalya/internal/restaurant"
	"bhojanalya/internal/storage"
	"bhojanalya/internal/competition"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// --------------------------------------------------
	// ENV
	// --------------------------------------------------
	if os.Getenv("APP_ENV") != "production" {
		_ = godotenv.Load()
	}

	required := []string{
		"JWT_SECRET",
		"DATABASE_URL",
		"GEMINI_API_KEY",
		"GEMINI_MODEL",
		"R2_ACCESS_KEY",
		"R2_SECRET_KEY",
		"R2_BUCKET_NAME",
		"R2_ENDPOINT",
	}

	for _, k := range required {
		if os.Getenv(k) == "" {
			log.Fatalf("❌ Missing env var: %s", k)
		}
	}

	log.Println("✅ Environment loaded")

	// --------------------------------------------------
	// DATABASE
	// --------------------------------------------------
	pgDB := db.ConnectPostgres()
	defer pgDB.Close()

	// --------------------------------------------------
	// GIN + CORS
	// --------------------------------------------------
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Authorization",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --------------------------------------------------
	// STORAGE (R2)
	// --------------------------------------------------
	r2Client, err := storage.NewR2Client(context.Background())
	if err != nil {
		log.Fatal("❌ R2 init failed:", err)
	}
	log.Println("✅ R2 client initialized")

	// --------------------------------------------------
	// AUTH
	// --------------------------------------------------
	userRepo := auth.NewPostgresUserRepository(pgDB)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	// --------------------------------------------------
	// PROTECTED TEST
	// --------------------------------------------------
	protected := r.Group("/protected")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/ping", func(c *gin.Context) {
			email, _ := c.Get("userEmail")
			c.JSON(200, gin.H{"email": email})
		})
	}

	// --------------------------------------------------
	// RESTAURANT
	// --------------------------------------------------
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

	// --------------------------------------------------
	// MENU
	// --------------------------------------------------
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

	// --------------------------------------------------
	// OCR + LLM SERVICES
	// --------------------------------------------------
	llmClient := llm.NewGeminiClient()
	ocrRepo := ocr.NewRepository(pgDB)

	ocrService := ocr.NewService(
		ocrRepo,
		r2Client,
		llmClient,
		menuService,
	)

	// --------------------------------------------------
	// 🚀 START WORKERS (CRITICAL FIX)
	// --------------------------------------------------
	log.Println("🚀 Starting OCR + LLM workers")

	log.Println("🚀 Starting OCR worker")
	go ocrService.RunOCRWorker()

	log.Println("🚀 Starting LLM parsing worker")
	go ocrService.RunLLMWorker()

	// --------------------------------------------------
	// DEBUG / HEALTH
	// --------------------------------------------------
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := pgDB.Ping(ctx)
		c.JSON(200, gin.H{
			"status":   "ok",
			"database": err == nil,
			"workers":  "running",
		})
	})

	r.POST("/debug/trigger-parse/:id", func(c *gin.Context) {
		var id int
		fmt.Sscanf(c.Param("id"), "%d", &id)

		var raw string
		err := pgDB.QueryRow(
			context.Background(),
			"SELECT raw_text FROM menu_uploads WHERE id=$1",
			id,
		).Scan(&raw)

		if err != nil || raw == "" {
			c.JSON(400, gin.H{"error": "no raw_text"})
			return
		}

		c.JSON(200, gin.H{"parsed": true})
	})


	// competitive insights routes would go here
	competitionService := competition.NewService(pgDB)
	competitionHandler := competition.NewHandler(competitionService)

	// Admin only
	admin.POST("/competition/recompute", competitionHandler.Recompute)

	// Public / restaurant preview
	r.GET("/competition/insights", competitionHandler.Get)


	// --------------------------------------------------
	// START SERVER
	// --------------------------------------------------
	log.Println("✅ API running on http://localhost:8000")
	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}
