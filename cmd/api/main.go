package main

import (
	"bhojanalya/internal/auth"
	"bhojanalya/internal/competition"
	"bhojanalya/internal/db"
	"bhojanalya/internal/deals"
	"bhojanalya/internal/llm"
	"bhojanalya/internal/menu"
	"bhojanalya/internal/middleware"
	"bhojanalya/internal/ocr"
	"bhojanalya/internal/restaurant"
	"bhojanalya/internal/storage"
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// ───────────────────────── ENV ─────────────────────────
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
		"R2_PUBLIC_BASE_URL",
	}

	for _, k := range required {
		if os.Getenv(k) == "" {
			log.Fatalf("❌ Missing env var: %s", k)
		}
	}

	// ───────────────────────── DB ─────────────────────────
	pgDB := db.ConnectPostgres()
	defer pgDB.Close()

	// ───────────────────────── GIN ─────────────────────────
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://bhojanalya.netlify.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ───────────────────────── STORAGE ─────────────────────────
	r2Client, err := storage.NewR2Client(context.Background())
	if err != nil {
		log.Fatal("❌ R2 init failed:", err)
	}

	// ───────────────────────── AUTH ─────────────────────────
	userRepo := auth.NewPostgresUserRepository(pgDB)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)

		protected := authGroup.Group("/protected")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/ping", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "pong"})
			})

			// ✅ ADD THESE TWO LINES
			protected.GET("/onboarding", authHandler.GetOnboardingStatus)
			protected.PATCH("/onboarding", authHandler.UpdateOnboardingStatus)
		}
	}

	// ───────────────────────── CORE REPOS ─────────────────────────
	restaurantRepo := restaurant.NewPostgresRepository(pgDB)
	menuRepo := menu.NewPostgresRepository(pgDB)
	competitionRepo := competition.NewRepository(pgDB)
	dealRepo := deals.NewRepository(pgDB)

	// ───────────────────────── SERVICES (ORDER MATTERS) ─────────────────────────
	menuService := menu.NewService(menuRepo, r2Client)

	restaurantService := restaurant.NewService(
		restaurantRepo,
		menuService,
		competitionRepo,
		r2Client,
	)

	dealService := deals.NewService(
		dealRepo,
		restaurantRepo,
		competitionRepo,
	)

	competitionService := competition.NewService(pgDB)

	// ───────────────────────── HANDLERS ─────────────────────────
	restaurantHandler := restaurant.NewHandler(restaurantService)
	menuHandler := menu.NewHandler(menuService)
	adminMenuHandler := menu.NewAdminHandler(menuService)
	dealHandler := deals.NewHandler(dealService)
	competitionHandler := competition.NewHandler(competitionService)

	// ───────────────────────── RESTAURANT ROUTES ─────────────────────────

	// 1. OPEN ROUTE: Only requires login. (The Service handles Admin vs Owner logic)
	r.GET("/restaurants/:id/preview", middleware.AuthMiddleware(), restaurantHandler.Preview)

	// 2. RESTRICTED ROUTES: Strictly requires the RESTAURANT role
	restaurants := r.Group("/restaurants")
	restaurants.Use(
		middleware.AuthMiddleware(),
		middleware.RequireRole("RESTAURANT"),
	)
	{
		restaurants.POST("", restaurantHandler.CreateRestaurant)
		restaurants.GET("/me", restaurantHandler.ListMyRestaurants)
		restaurants.POST("/:id/images", restaurantHandler.UploadImages)
	}

	// ───────────────────────── DEAL ROUTES ─────────────────────────
	dealsGroup := r.Group("/restaurants/:id/deals")
	dealsGroup.Use(
		middleware.AuthMiddleware(),
		middleware.RequireRole("RESTAURANT"),
	)
	{
		dealsGroup.GET("/suggestion", dealHandler.GetDealSuggestion())
		dealsGroup.POST("", dealHandler.CreateDeal())
		dealsGroup.GET("", dealHandler.GetRestaurantDeals())
	}

	deleteDeal := r.Group("/deals")
	deleteDeal.Use(
		middleware.AuthMiddleware(),
		middleware.RequireRole("RESTAURANT"),
	)
	{
		deleteDeal.DELETE("/:id", dealHandler.DeleteDeal())
	}

	// ───────────────────────── MENU ROUTES ─────────────────────────
	menus := r.Group("/menus")
	menus.Use(middleware.AuthMiddleware())
	{
		menus.POST("/upload", menuHandler.Upload)

		// ✅ STATUS POLLING (Feature-1)
		menus.GET("/:restaurant_id/status", menuHandler.GetMenuStatus)

		// ✅ RETRY FAILED MENU (Feature-2)
		menus.POST("/:restaurant_id/retry", menuHandler.RetryMenu)
	}

	// ───────────────────────── ADMIN ROUTES ─────────────────────────
	admin := r.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(),
		middleware.RequireRole("ADMIN"),
	)
	{
		// Restaurants
		admin.GET("/restaurants/approved", restaurantHandler.ListApprovedRestaurants)
		admin.GET("/restaurants/:id", restaurantHandler.GetAdminRestaurantDetails)
		admin.POST("/restaurants/:id/approve", restaurantHandler.ApproveRestaurant)

		// Menus
		admin.GET("/menus/pending", adminMenuHandler.PendingMenus)

		// Competition (manual fallback)
		admin.POST("/competition/recompute", competitionHandler.Recompute)
	}

	// ───────────────────────── PUBLIC ─────────────────────────
	r.GET("/competition/insights", competitionHandler.Get)

	// ───────────────────────── OCR + LLM WORKERS ─────────────────────────
	llmClient := llm.NewGeminiClient()
	ocrRepo := ocr.NewRepository(pgDB)

	ocrService := ocr.NewService(
		ocrRepo,
		r2Client,
		llmClient,
		menuService,
		competitionService,
	)

	go ocrService.RunOCRWorker()
	go ocrService.RunLLMWorker()

	// ───────────────────────── HEALTH ─────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ───────────────────────── START ─────────────────────────
	log.Println("🚀 API running at http://localhost:8000")
	r.Run(":8000")
}

// --------------------------------------------------
func mustHaveBinary(name string) {
	if _, err := exec.LookPath(name); err != nil {
		log.Fatalf("Required binary missing: %s", name)
	}
}
