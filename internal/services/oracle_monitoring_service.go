package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// OracleMonitoringService provides monitoring and analytics for oracle providers
type OracleMonitoringService struct {
	oracleRepo repository.OracleRepositoryInterface
}

// NewOracleMonitoringService creates a new oracle monitoring service
func NewOracleMonitoringService(oracleRepo repository.OracleRepositoryInterface) *OracleMonitoringService {
	return &OracleMonitoringService{
		oracleRepo: oracleRepo,
	}
}

// GetProviderMetrics retrieves performance metrics for an oracle provider
func (s *OracleMonitoringService) GetProviderMetrics(ctx context.Context, providerID uuid.UUID) (*models.OracleMetrics, error) {
	metrics, err := s.oracleRepo.GetOracleMetrics(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider metrics: %w", err)
	}
	return metrics, nil
}

// GetProviderReliabilityScore retrieves the reliability score for an oracle provider
func (s *OracleMonitoringService) GetProviderReliabilityScore(ctx context.Context, providerID uuid.UUID) (float64, error) {
	score, err := s.oracleRepo.GetOracleReliabilityScore(ctx, providerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get reliability score: %w", err)
	}
	return score, nil
}

// GetRequestStats retrieves statistics for oracle requests
func (s *OracleMonitoringService) GetRequestStats(ctx context.Context, providerID *uuid.UUID, startDate, endDate *time.Time) (map[models.RequestStatus]int64, error) {
	stats, err := s.oracleRepo.GetRequestStats(ctx, providerID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get request stats: %w", err)
	}
	return stats, nil
}

// GetProviderHealthStatus retrieves the health status for an oracle provider
func (s *OracleMonitoringService) GetProviderHealthStatus(ctx context.Context, providerID uuid.UUID) (*models.OracleStatus, error) {
	// Get recent request statistics to determine health
	stats, err := s.GetRequestStats(ctx, &providerID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}

	// Calculate error rate
	var totalRequests, failedRequests int64
	for status, count := range stats {
		totalRequests += count
		if status == models.RequestStatusFailed {
			failedRequests = count
		}
	}

	var errorRate float64
	if totalRequests > 0 {
		errorRate = float64(failedRequests) / float64(totalRequests)
	}

	// Get provider details
	providers, err := s.oracleRepo.ListOracleProviders(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	var provider *models.OracleProvider
	for _, p := range providers {
		if p.ID == providerID {
			provider = p
			break
		}
	}

	if provider == nil {
		return nil, fmt.Errorf("provider not found: %s", providerID.String())
	}

	status := &models.OracleStatus{
		IsHealthy:    errorRate < 0.1, // Consider healthy if error rate is less than 10%
		ErrorRate:    errorRate,
		LastChecked:  time.Now(),
		Capacity:     100,                // Placeholder
		Load:         int(totalRequests), // Placeholder
		Version:      "1.0.0",            // Placeholder
		Capabilities: provider.Capabilities,
		Reliability:  provider.Reliability,
	}

	return status, nil
}

// GetDashboardMetrics retrieves metrics for the oracle monitoring dashboard
func (s *OracleMonitoringService) GetDashboardMetrics(ctx context.Context) (*OracleDashboardMetrics, error) {
	// Get all providers
	providers, err := s.oracleRepo.ListOracleProviders(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	// Collect metrics for each provider
	providerMetrics := make([]*ProviderMetric, 0, len(providers))
	var totalRequests, successfulRequests, failedRequests int64

	for _, provider := range providers {
		// Get request stats for the last 24 hours
		endDate := time.Now()
		startDate := endDate.Add(-24 * time.Hour)

		stats, err := s.GetRequestStats(ctx, &provider.ID, &startDate, &endDate)
		if err != nil {
			log.Printf("Warning: Failed to get stats for provider %s: %v", provider.ID.String(), err)
			continue
		}

		// Calculate metrics
		var providerTotal, providerSuccessful, providerFailed int64
		for status, count := range stats {
			providerTotal += count
			switch status {
			case models.RequestStatusCompleted:
				providerSuccessful += count
			case models.RequestStatusFailed:
				providerFailed += count
			}
		}

		// Update totals
		totalRequests += providerTotal
		successfulRequests += providerSuccessful
		failedRequests += providerFailed

		// Calculate success rate
		var successRate float64
		if providerTotal > 0 {
			successRate = float64(providerSuccessful) / float64(providerTotal) * 100
		}

		providerMetrics = append(providerMetrics, &ProviderMetric{
			ProviderID:         provider.ID,
			ProviderName:       provider.Name,
			ProviderType:       string(provider.Type),
			IsActive:           provider.IsActive,
			Reliability:        provider.Reliability,
			TotalRequests:      providerTotal,
			SuccessfulRequests: providerSuccessful,
			FailedRequests:     providerFailed,
			SuccessRate:        successRate,
		})
	}

	// Calculate overall success rate
	var overallSuccessRate float64
	if totalRequests > 0 {
		overallSuccessRate = float64(successfulRequests) / float64(totalRequests) * 100
	}

	// Calculate error rate
	var errorRate float64
	if totalRequests > 0 {
		errorRate = float64(failedRequests) / float64(totalRequests) * 100
	}

	metrics := &OracleDashboardMetrics{
		TotalProviders:     int64(len(providers)),
		ActiveProviders:    s.countActiveProviders(providers),
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		OverallSuccessRate: overallSuccessRate,
		ErrorRate:          errorRate,
		ProviderMetrics:    providerMetrics,
		LastUpdated:        time.Now(),
	}

	return metrics, nil
}

// GetSLAMonitoring retrieves SLA monitoring data for oracle providers
func (s *OracleMonitoringService) GetSLAMonitoring(ctx context.Context) (*SLAMonitoringReport, error) {
	// Get all providers
	providers, err := s.oracleRepo.ListOracleProviders(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	// Time periods for analysis
	now := time.Now()
	lastHour := now.Add(-1 * time.Hour)
	lastDay := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	providerSLAs := make([]*ProviderSLA, 0, len(providers))

	for _, provider := range providers {
		// Get stats for different time periods
		hourlyStats, err := s.GetRequestStats(ctx, &provider.ID, &lastHour, &now)
		if err != nil {
			log.Printf("Warning: Failed to get hourly stats for provider %s: %v", provider.ID.String(), err)
			continue
		}

		dailyStats, err := s.GetRequestStats(ctx, &provider.ID, &lastDay, &now)
		if err != nil {
			log.Printf("Warning: Failed to get daily stats for provider %s: %v", provider.ID.String(), err)
			continue
		}

		weeklyStats, err := s.GetRequestStats(ctx, &provider.ID, &lastWeek, &now)
		if err != nil {
			log.Printf("Warning: Failed to get weekly stats for provider %s: %v", provider.ID.String(), err)
			continue
		}

		// Calculate SLA metrics
		hourlySLA := s.calculateSLAMetrics(hourlyStats)
		dailySLA := s.calculateSLAMetrics(dailyStats)
		weeklySLA := s.calculateSLAMetrics(weeklyStats)

		providerSLAs = append(providerSLAs, &ProviderSLA{
			ProviderID:   provider.ID,
			ProviderName: provider.Name,
			ProviderType: string(provider.Type),
			HourlySLA:    hourlySLA,
			DailySLA:     dailySLA,
			WeeklySLA:    weeklySLA,
			Reliability:  provider.Reliability,
		})
	}

	report := &SLAMonitoringReport{
		ReportGenerated: now,
		ProviderSLAs:    providerSLAs,
	}

	return report, nil
}

// GetCostAnalysis retrieves cost analysis for oracle usage
func (s *OracleMonitoringService) GetCostAnalysis(ctx context.Context) (*CostAnalysisReport, error) {
	// This is a placeholder implementation. In a real system, this would integrate
	// with billing systems to track costs associated with oracle usage.

	providers, err := s.oracleRepo.ListOracleProviders(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	providerCosts := make([]*ProviderCost, 0, len(providers))
	var totalCost float64

	for _, provider := range providers {
		// Get request count for cost calculation
		stats, err := s.GetRequestStats(ctx, &provider.ID, nil, nil)
		if err != nil {
			log.Printf("Warning: Failed to get stats for provider %s: %v", provider.ID.String(), err)
			continue
		}

		var requestCount int64
		for _, count := range stats {
			requestCount += count
		}

		// Calculate cost (placeholder logic)
		// In a real implementation, this would use actual pricing data
		cost := float64(requestCount) * 0.001 // $0.001 per request as example

		providerCosts = append(providerCosts, &ProviderCost{
			ProviderID:   provider.ID,
			ProviderName: provider.Name,
			ProviderType: string(provider.Type),
			RequestCount: requestCount,
			Cost:         cost,
		})

		totalCost += cost
	}

	report := &CostAnalysisReport{
		ReportGenerated: time.Now(),
		TotalCost:       totalCost,
		ProviderCosts:   providerCosts,
	}

	return report, nil
}

// Helper methods

func (s *OracleMonitoringService) countActiveProviders(providers []*models.OracleProvider) int64 {
	var count int64
	for _, provider := range providers {
		if provider.IsActive {
			count++
		}
	}
	return count
}

func (s *OracleMonitoringService) calculateSLAMetrics(stats map[models.RequestStatus]int64) *SLAMetrics {
	var total, successful, failed int64
	for status, count := range stats {
		total += count
		switch status {
		case models.RequestStatusCompleted:
			successful += count
		case models.RequestStatusFailed:
			failed += count
		}
	}

	var successRate, errorRate float64
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
		errorRate = float64(failed) / float64(total) * 100
	}

	return &SLAMetrics{
		TotalRequests:      total,
		SuccessfulRequests: successful,
		FailedRequests:     failed,
		SuccessRate:        successRate,
		ErrorRate:          errorRate,
	}
}

// Data structures for monitoring and analytics

// OracleDashboardMetrics represents metrics for the oracle monitoring dashboard
type OracleDashboardMetrics struct {
	TotalProviders     int64             `json:"total_providers"`
	ActiveProviders    int64             `json:"active_providers"`
	TotalRequests      int64             `json:"total_requests"`
	SuccessfulRequests int64             `json:"successful_requests"`
	FailedRequests     int64             `json:"failed_requests"`
	OverallSuccessRate float64           `json:"overall_success_rate"`
	ErrorRate          float64           `json:"error_rate"`
	ProviderMetrics    []*ProviderMetric `json:"provider_metrics"`
	LastUpdated        time.Time         `json:"last_updated"`
}

// ProviderMetric represents metrics for a single oracle provider
type ProviderMetric struct {
	ProviderID         uuid.UUID `json:"provider_id"`
	ProviderName       string    `json:"provider_name"`
	ProviderType       string    `json:"provider_type"`
	IsActive           bool      `json:"is_active"`
	Reliability        float64   `json:"reliability"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	SuccessRate        float64   `json:"success_rate"`
}

// SLAMonitoringReport represents an SLA monitoring report
type SLAMonitoringReport struct {
	ReportGenerated time.Time      `json:"report_generated"`
	ProviderSLAs    []*ProviderSLA `json:"provider_slas"`
}

// ProviderSLA represents SLA metrics for a single provider
type ProviderSLA struct {
	ProviderID   uuid.UUID   `json:"provider_id"`
	ProviderName string      `json:"provider_name"`
	ProviderType string      `json:"provider_type"`
	HourlySLA    *SLAMetrics `json:"hourly_sla"`
	DailySLA     *SLAMetrics `json:"daily_sla"`
	WeeklySLA    *SLAMetrics `json:"weekly_sla"`
	Reliability  float64     `json:"reliability"`
}

// SLAMetrics represents SLA metrics
type SLAMetrics struct {
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	SuccessRate        float64 `json:"success_rate"`
	ErrorRate          float64 `json:"error_rate"`
}

// CostAnalysisReport represents a cost analysis report
type CostAnalysisReport struct {
	ReportGenerated time.Time       `json:"report_generated"`
	TotalCost       float64         `json:"total_cost"`
	ProviderCosts   []*ProviderCost `json:"provider_costs"`
}

// ProviderCost represents cost data for a single provider
type ProviderCost struct {
	ProviderID   uuid.UUID `json:"provider_id"`
	ProviderName string    `json:"provider_name"`
	ProviderType string    `json:"provider_type"`
	RequestCount int64     `json:"request_count"`
	Cost         float64   `json:"cost"`
}
