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
	"fmt"
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

	// Verify required environment variables
	requiredEnvVars := []string{
		"JWT_SECRET",
		"DATABASE_URL",
		"GEMINI_API_KEY",
		"R2_ACCESS_KEY",
		"R2_SECRET_KEY",
		"R2_BUCKET_NAME",
		"R2_ENDPOINT",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("❌ Required environment variable not set: %s", envVar)
		}
	}

	log.Println("✅ Environment loaded")
	log.Printf("✅ GEMINI_API_KEY is set (length: %d)", len(os.Getenv("GEMINI_API_KEY")))

	// --------------------
	// DATABASE
	// --------------------
	pgDB := db.ConnectPostgres()
	defer pgDB.Close()

	// --------------------
	// CORS CONFIG
	// --------------------
	r := gin.Default()

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
	log.Println("✅ R2 client initialized")

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
	// LLM CLIENT
	// --------------------
	llmClient := llm.NewGeminiClient()
	if llmClient == nil {
		log.Fatal("❌ Failed to initialize Gemini client")
	}
	log.Println("✅ Gemini client initialized")

	// --------------------
	// OCR WORKER (BACKGROUND)
	// --------------------
	ocrRepo := ocr.NewRepository(pgDB)

	ocrService := ocr.NewService(
		ocrRepo,
		r2Client,
		llmClient,
		menuService,
	)

	// --------------------
	// START WORKERS
	// --------------------
	go func() {
		log.Println("🚀 Starting OCR worker...")
		// Give debug output first
		ocrService.DebugPipeline()
		
		// Start monitoring
		//go ocrService.MonitorPipeline()
		
		// Start the worker
		if err := ocrService.Start(); err != nil {
			log.Fatal("❌ OCR worker crashed:", err)
		}
	}()

	// --------------------
	// TEST ENDPOINTS (DEV ONLY)
	// --------------------
	r.POST("/api/llm/test-gemini", llm.TestGeminiHandler)
	
	// Add debug endpoint for OCR pipeline
	r.GET("/debug/ocr", func(c *gin.Context) {
		ocrService.DebugPipeline()
		c.JSON(200, gin.H{
			"message": "OCR debug executed, check logs",
			"time":    time.Now().Format(time.RFC3339),
		})
	})
	
	// Add endpoint to manually trigger parsing for a specific ID
	r.POST("/debug/trigger-parse/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		var id int
		fmt.Sscanf(idStr, "%d", &id)
		
		if id <= 0 {
			c.JSON(400, gin.H{"error": "Invalid ID"})
			return
		}
		
		// Get raw text from database
		var rawText string
		err := pgDB.QueryRow(context.Background(), 
			"SELECT raw_text FROM menu_uploads WHERE id = $1", id).Scan(&rawText)
		
		if err != nil {
			c.JSON(404, gin.H{"error": "Record not found", "details": err.Error()})
			return
		}
		
		if rawText == "" {
			c.JSON(400, gin.H{"error": "No OCR text available"})
			return
		}
		
		// Process directly
		rec := ocr.OCRRecord{ID: id, RawText: rawText}
		err = ocrService.ProcessOCR(rec)
		
		if err != nil {
			c.JSON(500, gin.H{
				"error":   "Processing failed",
				"details": err.Error(),
				"id":      id,
			})
			return
		}
		
		c.JSON(200, gin.H{
			"message": "Processing triggered successfully",
			"id":      id,
			"text_length": len(rawText),
		})
	})

	// --------------------
	// HEALTH CHECK
	// --------------------
	r.GET("/health", func(c *gin.Context) {
		// Check database connection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		var dbStatus string
		if err := pgDB.Ping(ctx); err != nil {
			dbStatus = "ERROR: " + err.Error()
		} else {
			dbStatus = "OK"
		}
		
		c.JSON(200, gin.H{
			"status":        "ok",
			"timestamp":     time.Now().Format(time.RFC3339),
			"database":      dbStatus,
			"workers":       "running",
			"gemini_api_key": len(os.Getenv("GEMINI_API_KEY")) > 0,
		})
	})

	// --------------------
	// START SERVER
	// --------------------
	log.Println("✅ Server starting on http://localhost:8000")
	log.Println("📊 Available debug endpoints:")
	log.Println("   GET  /debug/ocr           - Check OCR pipeline status")
	log.Println("   POST /debug/trigger-parse/:id - Manually trigger parsing")
	log.Println("   GET  /health             - Health check")
	log.Println("   POST /api/llm/test-gemini - Test Gemini API")
	
	if err := r.Run(":8000"); err != nil {
		log.Fatal("❌ Server failed:", err)
	}
}