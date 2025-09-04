package main

import (
	"fmt"
	"log"

	"github.com/smart-payment-infrastructure/internal/services"
)

func main() {
	// Example usage of ConsolidatedXRPLService

	// 1. Create service configuration
	config := services.ConsolidatedXRPLConfig{
		NetworkURL:   "https://s.altnet.rippletest.net:51234", // Testnet
		WebSocketURL: "wss://s.altnet.rippletest.net:51233",   // Testnet WebSocket
		TestNet:      true,
	}

	// 2. Create the service
	service := services.NewConsolidatedXRPLService(config)

	// 3. Initialize the service
	if err := service.Initialize(); err != nil {
		log.Fatalf("Failed to initialize XRPL service: %v", err)
	}

	// 4. Create a new wallet
	wallet, err := service.CreateWallet()
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}
	fmt.Printf("Created wallet: %s\n", wallet.Address)

	// 5. Get account information
	accountInfo, err := service.GetAccountInfo(wallet.Address)
	if err != nil {
		log.Printf("Failed to get account info: %v", err)
	} else {
		fmt.Printf("Account info: %+v\n", accountInfo)
	}

	// 6. Health check
	if err := service.HealthCheck(); err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("Service is healthy")
	}

	// 7. Example of creating an escrow with milestones (requires milestone data)
	fmt.Println("Service is ready for Smart Cheque operations")
	fmt.Println("Use CreateSmartChequeEscrowWithKey for escrow creation")
	fmt.Println("Use CompleteSmartChequeMilestoneWithKey for milestone completion")
}
