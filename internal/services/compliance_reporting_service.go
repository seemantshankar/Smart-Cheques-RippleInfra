package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
)

// ComplianceReportingService provides comprehensive compliance reporting functionality
type ComplianceReportingService struct {
	complianceRepo  repository.ComplianceRepositoryInterface
	transactionRepo repository.TransactionRepositoryInterface
	auditRepo       repository.AuditRepositoryInterface
}

// NewComplianceReportingService creates a new compliance reporting service
func NewComplianceReportingService(
	complianceRepo repository.ComplianceRepositoryInterface,
	transactionRepo repository.TransactionRepositoryInterface,
	auditRepo repository.AuditRepositoryInterface,
) *ComplianceReportingService {
	return &ComplianceReportingService{
		complianceRepo:  complianceRepo,
		transactionRepo: transactionRepo,
		auditRepo:       auditRepo,
	}
}

// GenerateComplianceReport generates a comprehensive compliance report
func (s *ComplianceReportingService) GenerateComplianceReport(
	ctx context.Context,
	enterpriseID string,
	reportType string,
	periodStart time.Time,
	periodEnd time.Time,
	generatedBy string,
) (*models.ComprehensiveComplianceReport, error) {
	// Get compliance statistics
	stats, err := s.complianceRepo.GetComplianceStats(&enterpriseID, &periodStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance stats: %w", err)
	}

	// Get flagged transactions
	flaggedTransactions, err := s.complianceRepo.GetComplianceStatusesByStatus("flagged", 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get flagged transactions: %w", err)
	}

	// Calculate compliance rate
	complianceRate := 1.0
	if stats.TotalTransactions > 0 {
		complianceRate = float64(stats.ApprovedTransactions) / float64(stats.TotalTransactions)
	}

	// Count violations by severity
	criticalViolations := int64(0)
	highViolations := int64(0)
	mediumViolations := int64(0)
	lowViolations := int64(0)

	// Analyze flagged transactions for violation details
	for _, tx := range flaggedTransactions {
		// This would typically analyze the actual violations
		// For now, we'll use a simple heuristic
		if len(tx.Violations) > 3 {
			criticalViolations++
		} else if len(tx.Violations) > 2 {
			highViolations++
		} else if len(tx.Violations) > 1 {
			mediumViolations++
		} else {
			lowViolations++
		}
	}

	// Calculate risk score and trend
	riskScore := s.calculateRiskScore(stats, flaggedTransactions)
	riskTrend := s.determineRiskTrend(enterpriseID, periodStart, periodEnd)

	// Generate recommendations
	recommendations := s.generateRecommendations(stats, flaggedTransactions, complianceRate)
	actionItems := s.generateActionItems(stats, flaggedTransactions)

	report := &models.ComprehensiveComplianceReport{
		ID:           uuid.New().String(),
		EnterpriseID: enterpriseID,
		ReportType:   reportType,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		GeneratedAt:  time.Now(),
		GeneratedBy:  generatedBy,
	}

	// Set calculated fields
	report.TotalTransactions = stats.TotalTransactions
	report.CompliantTransactions = stats.ApprovedTransactions
	report.NonCompliantTransactions = stats.FlaggedTransactions + stats.RejectedTransactions
	report.ComplianceRate = complianceRate
	report.TotalViolations = criticalViolations + highViolations + mediumViolations + lowViolations
	report.CriticalViolations = criticalViolations
	report.HighViolations = highViolations
	report.MediumViolations = mediumViolations
	report.LowViolations = lowViolations
	report.RiskScore = riskScore
	report.RiskTrend = riskTrend
	report.Recommendations = recommendations
	report.ActionItems = actionItems

	return report, nil
}

// GenerateComplianceTrend generates compliance trend analysis
func (s *ComplianceReportingService) GenerateComplianceTrend(
	ctx context.Context,
	enterpriseID string,
	periodStart time.Time,
	periodEnd time.Time,
	trendType string,
) (*models.ComplianceTrend, error) {
	// Generate data points for the trend
	dataPoints, err := s.generateTrendDataPoints(ctx, enterpriseID, periodStart, periodEnd, trendType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate trend data points: %w", err)
	}

	// Calculate trend direction and strength
	trendDirection, trendStrength := s.calculateTrendDirection(dataPoints)
	confidenceLevel := s.calculateConfidenceLevel(dataPoints)

	// Generate insights and anomalies
	keyInsights := s.generateKeyInsights(dataPoints, trendType)
	anomalies := s.detectAnomalies(dataPoints)
	predictions := s.generatePredictions(dataPoints, trendType)

	trend := &models.ComplianceTrend{
		EnterpriseID:    enterpriseID,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		TrendType:       trendType,
		DataPoints:      dataPoints,
		TrendDirection:  trendDirection,
		TrendStrength:   trendStrength,
		ConfidenceLevel: confidenceLevel,
		KeyInsights:     keyInsights,
		Anomalies:       anomalies,
		Predictions:     predictions,
	}

	return trend, nil
}

// GenerateComplianceDashboard generates dashboard data for compliance monitoring
func (s *ComplianceReportingService) GenerateComplianceDashboard(
	ctx context.Context,
	enterpriseID string,
) (map[string]interface{}, error) {
	// Get current period stats
	now := time.Now()
	periodStart := now.AddDate(0, 0, -30) // Last 30 days
	stats, err := s.complianceRepo.GetComplianceStats(&enterpriseID, &periodStart)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance stats: %w", err)
	}

	// Get recent flagged transactions
	flaggedTransactions, err := s.complianceRepo.GetComplianceStatusesByStatus("flagged", 10, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get flagged transactions: %w", err)
	}

	// Calculate key metrics
	complianceRate := 1.0
	if stats.TotalTransactions > 0 {
		complianceRate = float64(stats.ApprovedTransactions) / float64(stats.TotalTransactions)
	}

	riskScore := s.calculateRiskScore(stats, flaggedTransactions)

	// Generate dashboard data
	dashboard := map[string]interface{}{
		"enterprise_id":         enterpriseID,
		"compliance_rate":       complianceRate,
		"risk_score":            riskScore,
		"total_transactions":    stats.TotalTransactions,
		"approved_transactions": stats.ApprovedTransactions,
		"flagged_transactions":  stats.FlaggedTransactions,
		"rejected_transactions": stats.RejectedTransactions,
		"pending_transactions":  stats.PendingTransactions,
		"recent_violations":     len(flaggedTransactions),
		"last_updated":          now,
		"period_start":          periodStart,
		"period_end":            now,
	}

	return dashboard, nil
}

// calculateRiskScore calculates a risk score based on compliance data
func (s *ComplianceReportingService) calculateRiskScore(
	stats *models.ComplianceStats,
	flaggedTransactions []models.TransactionComplianceStatus,
) float64 {
	// Base risk score starts at 0.5
	riskScore := 0.5

	// Adjust based on compliance rate
	complianceRate := 1.0
	if stats.TotalTransactions > 0 {
		complianceRate = float64(stats.ApprovedTransactions) / float64(stats.TotalTransactions)
	}

	if complianceRate < 0.8 {
		riskScore += 0.3
	} else if complianceRate < 0.9 {
		riskScore += 0.1
	}

	// Adjust based on flagged transactions
	if len(flaggedTransactions) > 10 {
		riskScore += 0.2
	} else if len(flaggedTransactions) > 5 {
		riskScore += 0.1
	}

	// Adjust based on rejected transactions
	rejectionRate := 0.0
	if stats.TotalTransactions > 0 {
		rejectionRate = float64(stats.RejectedTransactions) / float64(stats.TotalTransactions)
	}

	if rejectionRate > 0.1 {
		riskScore += 0.2
	} else if rejectionRate > 0.05 {
		riskScore += 0.1
	}

	// Cap risk score at 1.0
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	return riskScore
}

// determineRiskTrend determines the risk trend direction
func (s *ComplianceReportingService) determineRiskTrend(
	_ string,
	_ time.Time,
	_ time.Time,
) string {
	// This would typically compare current period with previous period
	// For now, return a stable trend
	return "stable"
}

// generateRecommendations generates compliance recommendations
func (s *ComplianceReportingService) generateRecommendations(
	stats *models.ComplianceStats,
	flaggedTransactions []models.TransactionComplianceStatus,
	complianceRate float64,
) []string {
	recommendations := []string{}

	if complianceRate < 0.9 {
		recommendations = append(recommendations, "Implement additional compliance monitoring")
	}

	if len(flaggedTransactions) > 5 {
		recommendations = append(recommendations, "Review flagged transaction patterns")
	}

	if stats.RejectedTransactions > 0 {
		recommendations = append(recommendations, "Investigate rejected transaction causes")
	}

	if stats.PendingTransactions > 10 {
		recommendations = append(recommendations, "Reduce transaction processing delays")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Maintain current compliance practices")
	}

	return recommendations
}

// generateActionItems generates actionable items for compliance improvement
func (s *ComplianceReportingService) generateActionItems(
	stats *models.ComplianceStats,
	flaggedTransactions []models.TransactionComplianceStatus,
) []string {
	actionItems := []string{}

	if len(flaggedTransactions) > 0 {
		actionItems = append(actionItems, "Review and resolve flagged transactions within 24 hours")
	}

	if stats.RejectedTransactions > 0 {
		actionItems = append(actionItems, "Update transaction validation rules")
	}

	if stats.PendingTransactions > 10 {
		actionItems = append(actionItems, "Optimize transaction processing workflow")
	}

	return actionItems
}

// generateTrendDataPoints generates data points for trend analysis
func (s *ComplianceReportingService) generateTrendDataPoints(
	_ context.Context,
	enterpriseID string,
	periodStart time.Time,
	periodEnd time.Time,
	trendType string,
) ([]models.ComplianceDataPoint, error) {
	var dataPoints []models.ComplianceDataPoint

	// Generate daily data points
	current := periodStart
	for current.Before(periodEnd) {
		nextDay := current.AddDate(0, 0, 1)

		// Get stats for this day
		stats, err := s.complianceRepo.GetComplianceStats(&enterpriseID, &current)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for %s: %w", current.Format("2006-01-02"), err)
		}

		// Calculate value based on trend type
		var value float64
		switch trendType {
		case "compliance_rate":
			if stats.TotalTransactions > 0 {
				value = float64(stats.ApprovedTransactions) / float64(stats.TotalTransactions)
			} else {
				value = 1.0
			}
		case "violation_count":
			value = float64(stats.FlaggedTransactions + stats.RejectedTransactions)
		case "risk_score":
			value = s.calculateRiskScore(stats, []models.TransactionComplianceStatus{})
		default:
			value = 0.0
		}

		// Calculate baseline (simple moving average)
		baseline := value // Simplified baseline calculation

		// Determine if this is an anomaly
		isAnomaly := s.isAnomaly(value, baseline)

		dataPoint := models.ComplianceDataPoint{
			Timestamp:  current,
			Value:      value,
			Baseline:   baseline,
			IsAnomaly:  isAnomaly,
			Confidence: 0.8, // Simplified confidence calculation
		}

		dataPoints = append(dataPoints, dataPoint)
		current = nextDay
	}

	return dataPoints, nil
}

// calculateTrendDirection calculates the trend direction and strength
func (s *ComplianceReportingService) calculateTrendDirection(dataPoints []models.ComplianceDataPoint) (string, float64) {
	if len(dataPoints) < 2 {
		return "stable", 0.0
	}

	// Calculate trend using linear regression
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(dataPoints))

	for i, point := range dataPoints {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Determine direction
	var direction string
	if slope > 0.01 {
		direction = "up"
	} else if slope < -0.01 {
		direction = "down"
	} else {
		direction = "stable"
	}

	// Calculate strength (absolute value of slope, normalized)
	strength := math.Abs(slope)
	if strength > 1.0 {
		strength = 1.0
	}

	return direction, strength
}

// calculateConfidenceLevel calculates the confidence level of the trend analysis
func (s *ComplianceReportingService) calculateConfidenceLevel(dataPoints []models.ComplianceDataPoint) float64 {
	if len(dataPoints) < 3 {
		return 0.5
	}

	// Calculate confidence based on data consistency
	var values []float64
	for _, point := range dataPoints {
		values = append(values, point.Value)
	}

	// Calculate coefficient of variation
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))

	coefficientOfVariation := math.Sqrt(variance) / mean

	// Convert to confidence level (lower variation = higher confidence)
	confidence := 1.0 - math.Min(coefficientOfVariation, 1.0)
	return math.Max(confidence, 0.1) // Minimum 10% confidence
}

// generateKeyInsights generates key insights from the data
func (s *ComplianceReportingService) generateKeyInsights(dataPoints []models.ComplianceDataPoint, trendType string) []string {
	insights := []string{}

	if len(dataPoints) == 0 {
		return insights
	}

	// Find min and max values
	minValue := dataPoints[0].Value
	maxValue := dataPoints[0].Value
	for _, point := range dataPoints {
		if point.Value < minValue {
			minValue = point.Value
		}
		if point.Value > maxValue {
			maxValue = point.Value
		}
	}

	// Generate insights based on trend type
	switch trendType {
	case "compliance_rate":
		if maxValue-minValue > 0.1 {
			insights = append(insights, "Compliance rate shows significant variation")
		}
		if minValue < 0.8 {
			insights = append(insights, "Compliance rate dropped below 80% threshold")
		}
	case "violation_count":
		if maxValue > 10 {
			insights = append(insights, "High violation count detected")
		}
		if maxValue-minValue > 5 {
			insights = append(insights, "Violation count shows high variability")
		}
	case "risk_score":
		if maxValue > 0.7 {
			insights = append(insights, "Risk score exceeded 70% threshold")
		}
	}

	return insights
}

// detectAnomalies detects anomalies in the data points
func (s *ComplianceReportingService) detectAnomalies(dataPoints []models.ComplianceDataPoint) []string {
	anomalies := []string{}

	for _, point := range dataPoints {
		if point.IsAnomaly {
			anomalies = append(anomalies, fmt.Sprintf("Anomaly detected on %s: value %.2f",
				point.Timestamp.Format("2006-01-02"), point.Value))
		}
	}

	return anomalies
}

// generatePredictions generates predictions based on the trend data
func (s *ComplianceReportingService) generatePredictions(dataPoints []models.ComplianceDataPoint, trendType string) []string {
	predictions := []string{}

	if len(dataPoints) < 3 {
		return predictions
	}

	// Simple prediction based on recent trend
	recentValues := dataPoints[len(dataPoints)-3:]
	trend := recentValues[len(recentValues)-1].Value - recentValues[0].Value

	switch trendType {
	case "compliance_rate":
		if trend > 0.05 {
			predictions = append(predictions, "Compliance rate expected to improve")
		} else if trend < -0.05 {
			predictions = append(predictions, "Compliance rate expected to decline")
		} else {
			predictions = append(predictions, "Compliance rate expected to remain stable")
		}
	case "violation_count":
		if trend > 2 {
			predictions = append(predictions, "Violation count expected to increase")
		} else if trend < -2 {
			predictions = append(predictions, "Violation count expected to decrease")
		} else {
			predictions = append(predictions, "Violation count expected to remain stable")
		}
	case "risk_score":
		if trend > 0.1 {
			predictions = append(predictions, "Risk score expected to increase")
		} else if trend < -0.1 {
			predictions = append(predictions, "Risk score expected to decrease")
		} else {
			predictions = append(predictions, "Risk score expected to remain stable")
		}
	}

	return predictions
}

// isAnomaly determines if a value is an anomaly
func (s *ComplianceReportingService) isAnomaly(value, baseline float64) bool {
	// Simple anomaly detection: value deviates more than 20% from baseline
	deviation := math.Abs(value-baseline) / baseline
	return deviation > 0.2
}
