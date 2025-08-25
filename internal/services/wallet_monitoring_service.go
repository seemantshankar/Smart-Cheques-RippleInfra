package services

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// WalletMonitoringService handles wallet monitoring and health checks
type WalletMonitoringService struct {
	walletRepo  repository.WalletRepositoryInterface
	xrplService repository.XRPLServiceInterface
}

// WalletHealthStatus represents the health status of wallets
type WalletHealthStatus struct {
	TotalWallets       int                  `json:"total_wallets"`
	ActiveWallets      int                  `json:"active_wallets"`
	PendingWallets     int                  `json:"pending_wallets"`
	SuspendedWallets   int                  `json:"suspended_wallets"`
	WhitelistedWallets int                  `json:"whitelisted_wallets"`
	InactiveWallets    []WalletActivityInfo `json:"inactive_wallets"`
	RecentActivity     []WalletActivityInfo `json:"recent_activity"`
	Timestamp          time.Time            `json:"timestamp"`
}

// WalletActivityInfo provides information about wallet activity
type WalletActivityInfo struct {
	WalletID          uuid.UUID           `json:"wallet_id"`
	EnterpriseID      uuid.UUID           `json:"enterprise_id"`
	Address           string              `json:"address"`
	Status            models.WalletStatus `json:"status"`
	LastActivity      *time.Time          `json:"last_activity"`
	DaysSinceActivity *int                `json:"days_since_activity,omitempty"`
}

// NewWalletMonitoringService creates a new wallet monitoring service
func NewWalletMonitoringService(walletRepo repository.WalletRepositoryInterface, xrplService repository.XRPLServiceInterface) *WalletMonitoringService {
	return &WalletMonitoringService{
		walletRepo:  walletRepo,
		xrplService: xrplService,
	}
}

// GetWalletHealthStatus returns the overall health status of all wallets
func (s *WalletMonitoringService) GetWalletHealthStatus() (*WalletHealthStatus, error) {
	// Get all wallets
	allWallets, err := s.walletRepo.GetAllWallets()
	if err != nil {
		return nil, fmt.Errorf("failed to get all wallets: %w", err)
	}

	status := &WalletHealthStatus{
		TotalWallets:    len(allWallets),
		Timestamp:       time.Now(),
		InactiveWallets: []WalletActivityInfo{},
		RecentActivity:  []WalletActivityInfo{},
	}

	inactivityThreshold := time.Now().AddDate(0, 0, -30)    // 30 days
	recentActivityThreshold := time.Now().AddDate(0, 0, -7) // 7 days

	for _, wallet := range allWallets {
		// Count by status
		switch wallet.Status {
		case models.WalletStatusActive:
			status.ActiveWallets++
		case models.WalletStatusPending:
			status.PendingWallets++
		case models.WalletStatusSuspended:
			status.SuspendedWallets++
		}

		// Count whitelisted
		if wallet.IsWhitelisted {
			status.WhitelistedWallets++
		}

		// Check for inactive wallets
		if wallet.LastActivity == nil || wallet.LastActivity.Before(inactivityThreshold) {
			activityInfo := WalletActivityInfo{
				WalletID:     wallet.ID,
				EnterpriseID: wallet.EnterpriseID,
				Address:      wallet.Address,
				Status:       wallet.Status,
				LastActivity: wallet.LastActivity,
			}

			if wallet.LastActivity != nil {
				days := int(time.Since(*wallet.LastActivity).Hours() / 24)
				activityInfo.DaysSinceActivity = &days
			}

			status.InactiveWallets = append(status.InactiveWallets, activityInfo)
		}

		// Check for recent activity
		if wallet.LastActivity != nil && wallet.LastActivity.After(recentActivityThreshold) {
			activityInfo := WalletActivityInfo{
				WalletID:     wallet.ID,
				EnterpriseID: wallet.EnterpriseID,
				Address:      wallet.Address,
				Status:       wallet.Status,
				LastActivity: wallet.LastActivity,
			}

			status.RecentActivity = append(status.RecentActivity, activityInfo)
		}
	}

	return status, nil
}

// PerformWalletHealthCheck performs health checks on individual wallets
func (s *WalletMonitoringService) PerformWalletHealthCheck(walletID uuid.UUID) error {
	wallet, err := s.walletRepo.GetByID(walletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	// Validate address format
	if !s.xrplService.ValidateAddress(wallet.Address) {
		log.Printf("Wallet %s has invalid address: %s", walletID, wallet.Address)
		return fmt.Errorf("invalid wallet address: %s", wallet.Address)
	}

	// Check if the wallet exists on XRPL (optional, may fail for new wallets)
	_, err = s.xrplService.GetAccountInfo(wallet.Address)
	if err != nil {
		log.Printf("Warning: Could not retrieve account info for wallet %s: %v", wallet.Address, err)
		// Don't fail the health check for this, as new wallets may not be activated on XRPL yet
	}

	log.Printf("Wallet %s passed health check", walletID)
	return nil
}

// MonitorInactiveWallets identifies and optionally takes action on inactive wallets
func (s *WalletMonitoringService) MonitorInactiveWallets(inactiveDays int) ([]WalletActivityInfo, error) {
	threshold := time.Now().AddDate(0, 0, -inactiveDays)

	allWallets, err := s.walletRepo.GetAllWallets()
	if err != nil {
		return nil, fmt.Errorf("failed to get all wallets: %w", err)
	}

	var inactiveWallets []WalletActivityInfo

	for _, wallet := range allWallets {
		if wallet.Status != models.WalletStatusActive {
			continue // Only check active wallets
		}

		if wallet.LastActivity == nil || wallet.LastActivity.Before(threshold) {
			activityInfo := WalletActivityInfo{
				WalletID:     wallet.ID,
				EnterpriseID: wallet.EnterpriseID,
				Address:      wallet.Address,
				Status:       wallet.Status,
				LastActivity: wallet.LastActivity,
			}

			if wallet.LastActivity != nil {
				days := int(time.Since(*wallet.LastActivity).Hours() / 24)
				activityInfo.DaysSinceActivity = &days
			}

			inactiveWallets = append(inactiveWallets, activityInfo)
		}
	}

	if len(inactiveWallets) > 0 {
		log.Printf("Found %d inactive wallets (>%d days)", len(inactiveWallets), inactiveDays)
	}

	return inactiveWallets, nil
}

// GetWalletMetrics returns basic metrics about wallet usage
func (s *WalletMonitoringService) GetWalletMetrics() (map[string]interface{}, error) {
	status, err := s.GetWalletHealthStatus()
	if err != nil {
		return nil, err
	}

	metrics := map[string]interface{}{
		"total_wallets":          status.TotalWallets,
		"active_wallets":         status.ActiveWallets,
		"pending_wallets":        status.PendingWallets,
		"suspended_wallets":      status.SuspendedWallets,
		"whitelisted_wallets":    status.WhitelistedWallets,
		"inactive_wallets_count": len(status.InactiveWallets),
		"recent_activity_count":  len(status.RecentActivity),
		"active_percentage":      float64(status.ActiveWallets) / float64(status.TotalWallets) * 100,
		"whitelisted_percentage": float64(status.WhitelistedWallets) / float64(status.TotalWallets) * 100,
		"timestamp":              status.Timestamp,
	}

	return metrics, nil
}
