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

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"github.com/smart-payment-infrastructure/internal/handlers"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/internal/services"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

func main() {
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
	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/api/v1/oracle/providers", oracleHandler.RegisterProvider).Methods("POST")
	router.HandleFunc("/api/v1/oracle/providers/{id}", oracleHandler.GetProvider).Methods("GET")
	router.HandleFunc("/api/v1/oracle/providers/{id}", oracleHandler.UpdateProvider).Methods("PUT")
	router.HandleFunc("/api/v1/oracle/providers/{id}", oracleHandler.DeleteProvider).Methods("DELETE")
	router.HandleFunc("/api/v1/oracle/providers", oracleHandler.ListProviders).Methods("GET")
	router.HandleFunc("/api/v1/oracle/providers/active", oracleHandler.GetActiveProviders).Methods("GET")
	router.HandleFunc("/api/v1/oracle/providers/type/{type}", oracleHandler.GetProvidersByType).Methods("GET")
	router.HandleFunc("/api/v1/oracle/providers/{id}/health", oracleHandler.HealthCheck).Methods("GET")

	router.HandleFunc("/api/v1/oracle/requests/{id}", oracleHandler.GetRequest).Methods("GET")
	router.HandleFunc("/api/v1/oracle/verify", oracleHandler.VerifyMilestone).Methods("POST")
	router.HandleFunc("/api/v1/oracle/verify/{request_id}", oracleHandler.GetVerificationResult).Methods("GET")
	router.HandleFunc("/api/v1/oracle/proof/{request_id}", oracleHandler.GetProof).Methods("GET")

	router.HandleFunc("/api/v1/oracle/monitoring/dashboard", oracleHandler.GetDashboardMetrics).Methods("GET")
	router.HandleFunc("/api/v1/oracle/monitoring/sla", oracleHandler.GetSLAMonitoring).Methods("GET")
	router.HandleFunc("/api/v1/oracle/monitoring/costs", oracleHandler.GetCostAnalysis).Methods("GET")

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
