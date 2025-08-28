package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/smart-payment-infrastructure/internal/config"
	"github.com/smart-payment-infrastructure/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize XRPL service
	xrplService := services.NewXRPLService(services.XRPLConfig{
		NetworkURL: cfg.XRPL.NetworkURL,
		TestNet:    cfg.XRPL.TestNet,
	})

	// Initialize the service
	if err := xrplService.Initialize(); err != nil {
		log.Fatalf("Failed to initialize XRPL service: %v", err)
	}

	log.Println("XRPL service started successfully")

	// Test wallet generation
	wallet, err := xrplService.CreateWallet()
	if err != nil {
		log.Fatalf("Failed to create test wallet: %v", err)
	}

	log.Printf("Generated test wallet: %s", wallet.Address)

	// Test address validation
	if xrplService.ValidateAddress(wallet.Address) {
		log.Printf("Address validation successful for: %s", wallet.Address)
	} else {
		log.Printf("Address validation failed for: %s", wallet.Address)
		os.Exit(1)
	}

	// Test health check
	if err := xrplService.HealthCheck(); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	log.Println("All XRPL operations completed successfully")
}
