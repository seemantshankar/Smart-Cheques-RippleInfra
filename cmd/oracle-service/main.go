package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/smart-payment-infrastructure/internal/handlers"
	"github.com/smart-payment-infrastructure/internal/middleware"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

func main() {
	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Get database connection string from environment
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgresql://user:password@localhost:5432/smart_payment?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Successfully connected to database")

	// Initialize repositories
	oracleRepo := repository.NewOracleRepository(db)

	// Initialize messaging service
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	messagingService, err := messaging.NewService(redisURL, "", 0)
	if err != nil {
		log.Printf("Warning: Failed to connect to messaging service: %v", err)
		// Continue without messaging service for now
		messagingService = &messaging.Service{}
	}

	// Initialize services
	oracleService := services.NewOracleService(oracleRepo, messagingService)
	verificationService := services.NewOracleVerificationService(oracleService, oracleRepo, messagingService)
	monitoringService := services.NewOracleMonitoringService(oracleRepo)

	// Initialize handlers
	oracleHandler := handlers.NewOracleHandler(oracleService, verificationService, monitoringService)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(middleware.ErrorHandler())
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Register routes
	v1 := router.Group("/api/v1/oracle")
	{
		// Provider management
		v1.POST("/providers", oracleHandler.RegisterProvider)
		v1.GET("/providers/:id", oracleHandler.GetProvider)
		v1.PUT("/providers/:id", oracleHandler.UpdateProvider)
		v1.DELETE("/providers/:id", oracleHandler.DeleteProvider)
		v1.GET("/providers", oracleHandler.ListProviders)
		v1.GET("/providers/active", oracleHandler.GetActiveProviders)
		v1.GET("/providers/type/:type", oracleHandler.GetProvidersByType)
		v1.GET("/providers/:id/health", oracleHandler.HealthCheck)

		// Request management
		v1.GET("/requests/:id", oracleHandler.GetRequest)

		// Verification endpoints
		v1.POST("/verify", oracleHandler.VerifyMilestone)
		v1.GET("/verify/:request_id", oracleHandler.GetVerificationResult)
		v1.GET("/proof/:request_id", oracleHandler.GetProof)

		// Monitoring endpoints
		v1.GET("/monitoring/dashboard", oracleHandler.GetDashboardMetrics)
		v1.GET("/monitoring/sla", oracleHandler.GetSLAMonitoring)
		v1.GET("/monitoring/costs", oracleHandler.GetCostAnalysis)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Oracle service starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down oracle service...")

	// Shutdown server gracefully
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Oracle service exited")
}
