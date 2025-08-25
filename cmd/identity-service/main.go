package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/smart-payment-infrastructure/internal/config"
	"github.com/smart-payment-infrastructure/internal/handlers"
	"github.com/smart-payment-infrastructure/internal/middleware"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/auth"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.Database.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Parse JWT token durations
	accessTokenDuration, err := time.ParseDuration(cfg.JWT.AccessTokenDuration)
	if err != nil {
		log.Fatalf("Invalid access token duration: %v", err)
	}

	refreshTokenDuration, err := time.ParseDuration(cfg.JWT.RefreshTokenDuration)
	if err != nil {
		log.Fatalf("Invalid refresh token duration: %v", err)
	}

	// Initialize services
	jwtService := auth.NewJWTService(cfg.JWT.SecretKey, accessTokenDuration, refreshTokenDuration)
	userRepo := repository.NewUserRepository(db)
	enterpriseRepo := repository.NewEnterpriseRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	authService := services.NewAuthService(userRepo, jwtService)
	enterpriseService := services.NewEnterpriseService(enterpriseRepo)
	auditService := services.NewAuditService(auditRepo)
	authHandler := handlers.NewAuthHandler(authService)
	enterpriseHandler := handlers.NewEnterpriseHandler(enterpriseService)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Initialize messaging service
	messagingService, err := messaging.NewMessagingService(
		cfg.Redis.URL,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("Failed to initialize messaging service: %v", err)
	}
	defer messagingService.Close()

	// Subscribe to relevant events
	err = messagingService.SubscribeToEvent(messaging.EventTypeEnterpriseRegistered, handleEnterpriseRegistered)
	if err != nil {
		log.Printf("Failed to subscribe to enterprise registered events: %v", err)
	}

	r := gin.Default()
	
	// Add messaging middleware
	r.Use(middleware.MessagingMiddleware(messagingService))
	
	// Add audit middleware for authenticated routes
	r.Use(middleware.AuditMiddleware(auditService))
	
	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "identity-service",
			"status":  "healthy",
		})
	})

	r.GET("/health/messaging", middleware.MessagingHealthCheck)

	// Authentication endpoints
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", handlers.AuthMiddleware(authService), authHandler.Logout)
		auth.GET("/me", handlers.AuthMiddleware(authService), authHandler.Me)
	}

	// Protected endpoints
	protected := r.Group("/api")
	protected.Use(handlers.AuthMiddleware(authService))
	{
		// Enterprise endpoints with RBAC
		protected.POST("/enterprises", 
			middleware.RequirePermission(models.PermissionCreateEnterprise),
			enterpriseHandler.RegisterEnterprise)
		
		protected.GET("/enterprises/:id", 
			middleware.RequirePermission(models.PermissionReadEnterprise),
			middleware.EnterpriseOwnership(),
			enterpriseHandler.GetEnterprise)
		
		protected.PUT("/enterprises/:id/kyb-status", 
			middleware.RequirePermission(models.PermissionApproveKYB),
			enterpriseHandler.UpdateKYBStatus)
		
		protected.POST("/enterprises/:id/documents", 
			middleware.RequirePermission(models.PermissionUploadDocument),
			middleware.EnterpriseOwnership(),
			enterpriseHandler.UploadDocument)
		
		protected.GET("/enterprises/:id/documents", 
			middleware.RequirePermission(models.PermissionViewDocument),
			middleware.EnterpriseOwnership(),
			enterpriseHandler.GetDocuments)
		
		protected.PUT("/documents/:docId/verify", 
			middleware.RequirePermission(models.PermissionVerifyDocument),
			enterpriseHandler.VerifyDocument)
		
		protected.POST("/enterprises/:id/kyb-check", 
			middleware.RequirePermission(models.PermissionManageKYB),
			enterpriseHandler.PerformKYBCheck)
		
		protected.POST("/enterprises/:id/compliance-check", 
			middleware.RequirePermission(models.PermissionRunChecks),
			enterpriseHandler.PerformComplianceCheck)

		// Audit endpoints
		protected.GET("/audit-logs", 
			middleware.RequirePermission(models.PermissionViewAuditLogs),
			auditHandler.GetAuditLogs)
		
		protected.GET("/audit-logs/me", 
			auditHandler.GetUserAuditLogs)
		
		protected.GET("/enterprises/:id/audit-logs", 
			middleware.RequirePermission(models.PermissionViewAuditLogs),
			middleware.EnterpriseOwnership(),
			auditHandler.GetEnterpriseAuditLogs)
	}

	log.Println("Identity Service starting on :8001")
	log.Fatal(http.ListenAndServe(":8001", r))
}

func handleEnterpriseRegistered(event *messaging.Event) error {
	log.Printf("Handling enterprise registered event: %+v", event)
	// TODO: Implement actual business logic
	return nil
}

