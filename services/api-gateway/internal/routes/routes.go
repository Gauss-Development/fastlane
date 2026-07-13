package routes

import (
	"github.com/gin-gonic/gin"

	"api-gateway/internal/clients"
	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
	"api-gateway/pkg/metrics"
)

func SetupRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	searchHandler *handlers.SearchHandler,
	rfqHandler *handlers.RFQHandler,
	projectHandler *handlers.ProjectHandler,
	manufacturerHandler *handlers.ManufacturerHandler,
	healthHandler *handlers.HealthHandler,
	authClient *clients.AuthClient,
	redisClient *clients.RedisClient,
	cfg *config.Config,
) {
	// Health check route (no auth required)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/metrics", gin.WrapH(metrics.Handler()))

	// Global middleware
	router.Use(middleware.RequestValidator(cfg.RequestMaxBodyBytes))
	router.Use(middleware.RateLimit(redisClient, cfg.RateLimit))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (no authentication required)
		authGroup := v1.Group("/auth")
		{
			// Email/password
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)

			// OAuth2 flow
			authGroup.GET("/google", authHandler.GetGoogleAuthURL)
			authGroup.GET("/google/callback", authHandler.GoogleCallback)
			authGroup.POST("/exchange", authHandler.ExchangeAuthCode)

			// Token management
			authGroup.POST("/refresh", authHandler.RefreshToken)

			// Protected auth routes
			authProtected := authGroup.Group("")
			authProtected.Use(middleware.AuthMiddleware(authClient))
			{
				authProtected.POST("/logout", authHandler.Logout)
				authProtected.GET("/validate", authHandler.ValidateToken)
			}
		}

		// Public routes (no authentication required)
		publicGroup := v1.Group("/public")
		publicGroup.Use(middleware.OptionalAuthMiddleware(authClient))
		{
			// Public user routes
			publicUsers := publicGroup.Group("/users")
			{
				publicUsers.GET("/search", userHandler.SearchUsers)
				publicUsers.GET("/stats", userHandler.GetStats)
				publicUsers.GET("/:id/profile", userHandler.GetUserProfile)
			}
		}

		// Supplier magic-link RFQ surface: public, gated by the signed
		// token in the path, not by a session (suppliers have no login).
		supplierRFQ := v1.Group("/supplier-rfq")
		{
			supplierRFQ.GET("/:token", rfqHandler.SupplierGetRFQ)
			supplierRFQ.POST("/:token/quote", rfqHandler.SupplierSubmitQuote)
		}

		// Protected routes (authentication required)
		protectedGroup := v1.Group("")
		protectedGroup.Use(middleware.AuthMiddleware(authClient))
		{
			// Hybrid AI product search (Claude spec extraction + pgvector ranking)
			protectedGroup.POST("/search", searchHandler.Search)

			// RFQ flow: buyer creates and tracks quote requests
			rfqs := protectedGroup.Group("/rfqs")
			{
				rfqs.POST("", rfqHandler.CreateRFQ)
				rfqs.GET("", rfqHandler.ListRFQs)
				rfqs.GET("/:id", rfqHandler.GetRFQ)
				rfqs.GET("/:id/quotes", rfqHandler.ListQuotes)
			}

			// Design flow: startup projects, design files, and per-manufacturer NDAs.
			// gateway maps userID -> owner_id / actor_id / manufacturer_id.
			projects := protectedGroup.Group("/projects")
			{
				projects.POST("", middleware.RequireRole("startup"), projectHandler.CreateProject)
				projects.GET("", projectHandler.ListProjects)
				projects.GET("/:id", projectHandler.GetProject)
				projects.POST("/:id/files/upload-url", projectHandler.RequestUploadURL)
				projects.GET("/:id/files", projectHandler.ListFiles)
				projects.POST("/:id/invite", middleware.RequireRole("startup"), projectHandler.InviteManufacturer)
				projects.POST("/:id/nda/accept", middleware.RequireRole("manufacturer"), projectHandler.AcceptNDA)
				projects.GET("/:id/nda", projectHandler.GetNDAStatus)
			}

			files := protectedGroup.Group("/files")
			{
				files.POST("/:fileId/confirm", projectHandler.ConfirmUpload)
				files.GET("/:fileId/download-url", projectHandler.RequestDownloadURL)
			}

			// Catalog flow: Chinese PCB/PCBA manufacturer profiles + capabilities.
			// gateway maps userID -> manufacturer.user_id / actor_id.
			manuf := protectedGroup.Group("/manufacturers")
			{
				manuf.POST("", middleware.RequireRole("manufacturer"), manufacturerHandler.CreateManufacturer)
				manuf.GET("", manufacturerHandler.ListManufacturers)
				manuf.GET("/:id", manufacturerHandler.GetManufacturer)
				manuf.PUT("/:id", middleware.RequireRole("manufacturer"), manufacturerHandler.UpdateManufacturer)
				manuf.POST("/:id/verify", middleware.RequireRole("admin"), manufacturerHandler.VerifyManufacturer)
			}
			// Separate top-level path (not under /manufacturers) to avoid gin
			// static-vs-:id sibling conflicts.
			protectedGroup.GET("/manufacturer-profile", manufacturerHandler.GetMyManufacturer)

			// User routes
			users := protectedGroup.Group("/users")
			{
				users.POST("", userHandler.CreateUser)
				users.GET("", userHandler.ListUsers)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
			}
		}
	}
}
