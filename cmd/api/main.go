package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
	"bhojanalya/internal/deals"
	"bhojanalya/internal/auth"
	"bhojanalya/internal/competition"
	"bhojanalya/internal/db"
	"bhojanalya/internal/llm"
	"bhojanalya/internal/menu"
	"bhojanalya/internal/middleware"
	"bhojanalya/internal/ocr"
	"bhojanalya/internal/restaurant"
	"bhojanalya/internal/storage"
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
		// "GEMINI_MODEL", // Optional, has default
		"R2_ACCESS_KEY",
		"R2_SECRET_KEY",
		"R2_BUCKET_NAME",
		"R2_ENDPOINT",
	}

	for _, k := range required {
		if os.Getenv(k) == "" {
			log.Printf("⚠️  Missing env var: %s", k)
			// Don't fatal for GEMINI_MODEL if it has default
			if k != "GEMINI_MODEL" {
				log.Fatalf("❌ Missing required env var: %s", k)
			}
		}
	}

	log.Println("✅ Environment loaded")
	log.Printf("✅ GEMINI_API_KEY present: %v", os.Getenv("GEMINI_API_KEY") != "")
	log.Printf("✅ R2 credentials present: %v", os.Getenv("R2_ACCESS_KEY_ID") != "")

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
		ExposeHeaders: []string{
			"Content-Length",
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
// --------------------------------------------------
// RESTAURANT
// --------------------------------------------------
	restaurantRepo := restaurant.NewPostgresRepository(pgDB)

	// competition repo (READ ONLY)
	competitionRepo := competition.NewRepository(pgDB)

	// restaurant service
	restaurantService := restaurant.NewService(
		restaurantRepo,
		competitionRepo,
	)

	restaurantHandler := restaurant.NewHandler(restaurantService)

	// --------------------------------------------------
	// DEALS
	// --------------------------------------------------
	dealService := deals.NewService(
		restaurantRepo, // implements core.RestaurantReader
		competitionRepo,
	)

	dealHandler := deals.NewHandler(dealService)

	// --------------------------------------------------
	// ROUTES
	// --------------------------------------------------
	restaurantRoutes := r.Group("/restaurants")
	restaurantRoutes.Use(
		middleware.AuthMiddleware(),
		middleware.RequireRole("RESTAURANT"),
	)
	{
		restaurantRoutes.POST("", restaurantHandler.CreateRestaurant)
		restaurantRoutes.GET("/me", restaurantHandler.ListMyRestaurants)

		// 🔥 DEAL SUGGESTIONS
		restaurantRoutes.GET(
			"/:id/deals",
			dealHandler.GetDealSuggestion(),
		)
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
	if llmClient == nil {
		log.Fatal("❌ Failed to initialize Gemini client")
	}
	log.Println("✅ Gemini client initialized")

	ocrRepo := ocr.NewRepository(pgDB)

	ocrService := ocr.NewService(
		ocrRepo,
		r2Client,
		llmClient,
		menuService,
	)

	// --------------------------------------------------
	// COMPETITION SERVICE
	// --------------------------------------------------
	competitionService := competition.NewService(pgDB)
	competitionHandler := competition.NewHandler(competitionService)

	// Admin only
	admin.POST("/competition/recompute", competitionHandler.Recompute)

	// Public / restaurant preview
	r.GET("/competition/insights", competitionHandler.Get)

	// --------------------------------------------------
	// 🚀 START WORKERS (CRITICAL FIX)
	// --------------------------------------------------
	log.Println("🚀 Starting OCR + LLM workers")

	go func() {
		log.Println("🚀 Starting OCR worker")
		ocrService.RunOCRWorker()
	}()

	go func() {
		log.Println("🚀 Starting LLM parsing worker")
		ocrService.RunLLMWorker()
	}()

	// --------------------------------------------------
	// DEBUG / HEALTH
	// --------------------------------------------------
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := pgDB.Ping(ctx); err != nil {
			dbStatus = err.Error()
		}

		c.JSON(200, gin.H{
			"status":        "ok",
			"database":      dbStatus,
			"workers":       "running",
			"timestamp":     time.Now().Format(time.RFC3339),
			"gemini_key":    os.Getenv("GEMINI_API_KEY") != "",
			"r2_configured": os.Getenv("R2_ACCESS_KEY_ID") != "",
		})
	})

	r.POST("/debug/trigger-parse/:id", func(c *gin.Context) {
		var id int
		fmt.Sscanf(c.Param("id"), "%d", &id)

		var raw string
		queryErr := pgDB.QueryRow(
			context.Background(),
			"SELECT raw_text FROM menu_uploads WHERE id=$1",
			id,
		).Scan(&raw)

		if queryErr != nil || raw == "" {
			c.JSON(400, gin.H{"error": "no raw_text found for ID", "id": id})
			return
		}

		c.JSON(200, gin.H{
			"parsed":      true,
			"id":          id,
			"text_length": len(raw),
			"message":     "Raw text exists (parsing triggered by worker)",
		})
	})

	// Add debug endpoint for OCR status
	r.GET("/debug/ocr-status", func(c *gin.Context) {
		// Check database status
		var pendingCount int
		queryErr := pgDB.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM menu_uploads WHERE status = 'MENU_UPLOADED'").Scan(&pendingCount)
		
		if queryErr != nil {
			c.JSON(500, gin.H{"error": queryErr.Error()})
			return
		}

		c.JSON(200, gin.H{
			"pending_uploads": pendingCount,
			"message": "Check logs for worker activity",
			"help": "Upload a menu with restaurant_id in form data",
		})
	})

	// --------------------------------------------------
	// START SERVER
	// --------------------------------------------------
	log.Println("✅ API running on http://localhost:8000")
	log.Println("📋 Available endpoints:")
	log.Println("   POST /menus/upload            - Upload menu (include restaurant_id)")
	log.Println("   GET  /health                  - Health check")
	log.Println("   GET  /debug/ocr-status        - Check OCR worker status")
	log.Println("   POST /debug/trigger-parse/:id - Manually trigger parsing")
	log.Println("")
	log.Println("⚠️  IMPORTANT: Upload must include 'restaurant_id' in form data")
	
	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}
func mustHaveBinary(name string) {
	if _, err := exec.LookPath(name); err != nil {
		log.Fatalf("Required binary not found in PATH: %s", name)
	}
}
