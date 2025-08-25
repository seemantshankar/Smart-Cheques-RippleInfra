package services

import (
	"fmt"
	"log"

	"github.com/smart-payment-infrastructure/pkg/xrpl"
)

type XRPLService struct {
	client      *xrpl.Client
	initialized bool
}

type XRPLConfig struct {
	NetworkURL string
	TestNet    bool
}

func NewXRPLService(config XRPLConfig) *XRPLService {
	client := xrpl.NewClient(config.NetworkURL, config.TestNet)
	return &XRPLService{
		client: client,
	}
}

func (s *XRPLService) Initialize() error {
	if err := s.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to XRPL: %w", err)
	}

	if err := s.client.HealthCheck(); err != nil {
		return fmt.Errorf("XRPL health check failed: %w", err)
	}

	s.initialized = true
	log.Println("XRPL service initialized successfully")
	return nil
}

func (s *XRPLService) CreateWallet() (*xrpl.WalletInfo, error) {
	wallet, err := s.client.GenerateWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	log.Printf("Created new XRPL wallet: %s", wallet.Address)
	return wallet, nil
}

func (s *XRPLService) ValidateAddress(address string) bool {
	return s.client.ValidateAddress(address)
}

func (s *XRPLService) GetAccountInfo(address string) (interface{}, error) {
	accountInfo, err := s.client.GetAccountInfo(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info for %s: %w", address, err)
	}

	return accountInfo, nil
}

func (s *XRPLService) HealthCheck() error {
	if !s.initialized {
		return fmt.Errorf("XRPL service not initialized")
	}
	return s.client.HealthCheck()
}