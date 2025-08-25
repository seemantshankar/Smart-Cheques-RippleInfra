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
	walletRepo := repository.NewWalletRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Initialize XRPL service
	xrplService := services.NewXRPLService(services.XRPLConfig{
		NetworkURL: cfg.XRPL.NetworkURL,
		TestNet:    cfg.XRPL.TestNet,
	})

	if err := xrplService.Initialize(); err != nil {
		log.Fatalf("Failed to initialize XRPL service: %v", err)
	}

	// Initialize wallet service
	walletService, err := services.NewWalletService(
		walletRepo,
		enterpriseRepo,
		xrplService,
		services.WalletServiceConfig{
			EncryptionKey: cfg.JWT.SecretKey, // Using JWT secret as encryption key for now
		},
	)
	if err != nil {
		log.Fatalf("Failed to initialize wallet service: %v", err)
	}

	// Initialize wallet monitoring service
	walletMonitoringService := services.NewWalletMonitoringService(walletRepo, xrplService)

	authService := services.NewAuthService(userRepo, jwtService)
	enterpriseService := services.NewEnterpriseService(enterpriseRepo, walletService)
	auditService := services.NewAuditService(auditRepo)
	authHandler := handlers.NewAuthHandler(authService)
	enterpriseHandler := handlers.NewEnterpriseHandler(enterpriseService)
	walletHandler := handlers.NewWalletHandler(walletService, walletMonitoringService)
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

		// Wallet endpoints with RBAC
		protected.POST("/wallets",
			middleware.RequirePermission(models.PermissionCreateEnterprise),
			walletHandler.CreateWallet)

		protected.GET("/wallets/:id",
			middleware.RequirePermission(models.PermissionReadEnterprise),
			walletHandler.GetWallet)

		protected.GET("/wallets/address/:address",
			middleware.RequirePermission(models.PermissionReadEnterprise),
			walletHandler.GetWalletByAddress)

		protected.GET("/enterprises/:enterpriseId/wallets",
			middleware.RequirePermission(models.PermissionReadEnterprise),
			middleware.EnterpriseOwnership(),
			walletHandler.GetEnterpriseWallets)

		protected.PUT("/wallets/:id/activate",
			middleware.RequirePermission(models.PermissionApproveKYB),
			walletHandler.ActivateWallet)

		protected.PUT("/wallets/:id/whitelist",
			middleware.RequirePermission(models.PermissionApproveKYB),
			walletHandler.WhitelistWallet)

		protected.PUT("/wallets/:id/suspend",
			middleware.RequirePermission(models.PermissionApproveKYB),
			walletHandler.SuspendWallet)

		protected.GET("/wallets/whitelisted",
			middleware.RequirePermission(models.PermissionViewAuditLogs),
			walletHandler.GetWhitelistedWallets)

		protected.POST("/wallets/validate-address",
			middleware.RequirePermission(models.PermissionReadEnterprise),
			walletHandler.ValidateAddress)

		// Wallet monitoring endpoints
		protected.GET("/wallets/health-status",
			middleware.RequirePermission(models.PermissionViewAuditLogs),
			walletHandler.GetWalletHealthStatus)

		protected.POST("/wallets/:id/health-check",
			middleware.RequirePermission(models.PermissionApproveKYB),
			walletHandler.PerformWalletHealthCheck)

		protected.GET("/wallets/inactive",
			middleware.RequirePermission(models.PermissionViewAuditLogs),
			walletHandler.GetInactiveWallets)

		protected.GET("/wallets/metrics",
			middleware.RequirePermission(models.PermissionViewAuditLogs),
			walletHandler.GetWalletMetrics)
	}

	log.Println("Identity Service starting on :8001")
	log.Fatal(http.ListenAndServe(":8001", r))
}

func handleEnterpriseRegistered(event *messaging.Event) error {
	log.Printf("Handling enterprise registered event: %+v", event)
	// TODO: Implement actual business logic
	return nil
}
