package services

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/crypto"
)

type WalletService struct {
	walletRepo      *repository.WalletRepository
	enterpriseRepo  *repository.EnterpriseRepository
	xrplService     *XRPLService
	encryptor       *crypto.Encryptor
}

type WalletServiceConfig struct {
	EncryptionKey string
}

func NewWalletService(
	walletRepo *repository.WalletRepository,
	enterpriseRepo *repository.EnterpriseRepository,
	xrplService *XRPLService,
	config WalletServiceConfig,
) (*WalletService, error) {
	encryptor, err := crypto.NewEncryptor(config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	return &WalletService{
		walletRepo:     walletRepo,
		enterpriseRepo: enterpriseRepo,
		xrplService:    xrplService,
		encryptor:      encryptor,
	}, nil
}

func (s *WalletService) CreateWalletForEnterprise(enterpriseID uuid.UUID, networkType string) (*models.WalletResponse, error) {
	// Verify enterprise exists
	enterprise, err := s.enterpriseRepo.GetByID(enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enterprise: %w", err)
	}

	// Check if enterprise already has an active wallet for this network
	existingWallet, err := s.walletRepo.GetActiveByEnterpriseAndNetwork(enterpriseID, networkType)
	if err == nil && existingWallet != nil {
		return nil, fmt.Errorf("enterprise already has an active wallet for network: %s", networkType)
	}

	// Generate XRPL wallet
	xrplWallet, err := s.xrplService.CreateWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create XRPL wallet: %w", err)
	}

	// Encrypt sensitive data
	encryptedPrivateKey, err := s.encryptor.Encrypt(xrplWallet.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	encryptedSeed, err := s.encryptor.Encrypt(xrplWallet.Seed)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt seed: %w", err)
	}

	// Create wallet record
	wallet := &models.Wallet{
		ID:                  uuid.New(),
		EnterpriseID:        enterpriseID,
		Address:             xrplWallet.Address,
		PublicKey:           xrplWallet.PublicKey,
		EncryptedPrivateKey: encryptedPrivateKey,
		EncryptedSeed:       encryptedSeed,
		Status:              models.WalletStatusPending,
		IsWhitelisted:       false,
		NetworkType:         networkType,
		Metadata:            make(map[string]string),
	}

	// Store in database
	if err := s.walletRepo.Create(wallet); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	log.Printf("Created wallet %s for enterprise %s (%s) on network %s", 
		wallet.Address, enterprise.LegalName, enterpriseID, networkType)

	return wallet.ToResponse(), nil
}

func (s *WalletService) ActivateWallet(walletID uuid.UUID) error {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	if wallet.Status == models.WalletStatusActive {
		return fmt.Errorf("wallet is already active")
	}

	// Deactivate any existing active wallet for the same enterprise and network
	existingWallet, err := s.walletRepo.GetActiveByEnterpriseAndNetwork(wallet.EnterpriseID, wallet.NetworkType)
	if err == nil && existingWallet != nil && existingWallet.ID != walletID {
		existingWallet.Status = models.WalletStatusDeactivated
		if err := s.walletRepo.Update(existingWallet); err != nil {
			return fmt.Errorf("failed to deactivate existing wallet: %w", err)
		}
	}

	wallet.Status = models.WalletStatusActive
	if err := s.walletRepo.Update(wallet); err != nil {
		return fmt.Errorf("failed to activate wallet: %w", err)
	}

	log.Printf("Activated wallet %s for enterprise %s", wallet.Address, wallet.EnterpriseID)
	return nil
}

func (s *WalletService) WhitelistWallet(walletID uuid.UUID) error {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	if !wallet.IsActive() {
		return fmt.Errorf("cannot whitelist inactive wallet")
	}

	wallet.IsWhitelisted = true
	if err := s.walletRepo.Update(wallet); err != nil {
		return fmt.Errorf("failed to whitelist wallet: %w", err)
	}

	log.Printf("Whitelisted wallet %s", wallet.Address)
	return nil
}

func (s *WalletService) SuspendWallet(walletID uuid.UUID, reason string) error {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	wallet.Status = models.WalletStatusSuspended
	if wallet.Metadata == nil {
		wallet.Metadata = make(map[string]string)
	}
	wallet.Metadata["suspension_reason"] = reason
	wallet.Metadata["suspended_at"] = time.Now().Format(time.RFC3339)

	if err := s.walletRepo.Update(wallet); err != nil {
		return fmt.Errorf("failed to suspend wallet: %w", err)
	}

	log.Printf("Suspended wallet %s: %s", wallet.Address, reason)
	return nil
}

func (s *WalletService) GetWalletByID(walletID uuid.UUID) (*models.WalletResponse, error) {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet.ToResponse(), nil
}

func (s *WalletService) GetWalletByAddress(address string) (*models.WalletResponse, error) {
	wallet, err := s.walletRepo.GetByAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet.ToResponse(), nil
}

func (s *WalletService) GetWalletsForEnterprise(enterpriseID uuid.UUID) ([]*models.WalletResponse, error) {
	wallets, err := s.walletRepo.GetByEnterpriseID(enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %w", err)
	}

	responses := make([]*models.WalletResponse, len(wallets))
	for i, wallet := range wallets {
		responses[i] = wallet.ToResponse()
	}

	return responses, nil
}

func (s *WalletService) GetActiveWalletForEnterprise(enterpriseID uuid.UUID, networkType string) (*models.WalletResponse, error) {
	wallet, err := s.walletRepo.GetActiveByEnterpriseAndNetwork(enterpriseID, networkType)
	if err != nil {
		return nil, fmt.Errorf("failed to get active wallet: %w", err)
	}

	return wallet.ToResponse(), nil
}

func (s *WalletService) UpdateWalletActivity(walletID uuid.UUID) error {
	if err := s.walletRepo.UpdateLastActivity(walletID); err != nil {
		return fmt.Errorf("failed to update wallet activity: %w", err)
	}

	return nil
}

func (s *WalletService) GetWhitelistedWallets() ([]*models.WalletResponse, error) {
	wallets, err := s.walletRepo.GetWhitelistedWallets()
	if err != nil {
		return nil, fmt.Errorf("failed to get whitelisted wallets: %w", err)
	}

	responses := make([]*models.WalletResponse, len(wallets))
	for i, wallet := range wallets {
		responses[i] = wallet.ToResponse()
	}

	return responses, nil
}

func (s *WalletService) ValidateWalletAddress(address string) bool {
	return s.xrplService.ValidateAddress(address)
}

// GetDecryptedPrivateKey returns the decrypted private key for a wallet
// This should only be used for transaction signing and should be handled with extreme care
func (s *WalletService) GetDecryptedPrivateKey(walletID uuid.UUID) (string, error) {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return "", fmt.Errorf("failed to get wallet: %w", err)
	}

	if !wallet.CanTransact() {
		return "", fmt.Errorf("wallet cannot transact (status: %s, whitelisted: %v)", wallet.Status, wallet.IsWhitelisted)
	}

	privateKey, err := s.encryptor.Decrypt(wallet.EncryptedPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return privateKey, nil
}