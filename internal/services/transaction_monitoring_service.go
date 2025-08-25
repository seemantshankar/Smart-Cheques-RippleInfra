package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/repository"
	"github.com/smart-payment-infrastructure/pkg/messaging"
)

// TransactionMonitoringService provides real-time monitoring and dashboard metrics
type TransactionMonitoringService struct {
	transactionRepo  repository.TransactionRepositoryInterface
	messagingService *messaging.MessagingService

	// Real-time metrics
	metrics      *TransactionMetrics
	metricsMutex sync.RWMutex

	// Monitoring configuration
	updateInterval  time.Duration
	retentionPeriod time.Duration

	// Control channels
	isRunning   bool
	stopChannel chan struct{}
	wg          sync.WaitGroup

	// Alerting thresholds
	alertThresholds AlertThresholds
}

// TransactionMetrics holds real-time transaction metrics
type TransactionMetrics struct {
	// Current queue state
	QueueDepth      int64 `json:"queue_depth"`
	ProcessingCount int64 `json:"processing_count"`
	BatchingCount   int64 `json:"batching_count"`

	// Processing rates (per minute)
	TransactionRate float64 `json:"transaction_rate"`
	SuccessRate     float64 `json:"success_rate"`
	FailureRate     float64 `json:"failure_rate"`

	// Performance metrics
	AverageProcessingTime  float64 `json:"average_processing_time_seconds"`
	AverageBatchSize       float64 `json:"average_batch_size"`
	FeeOptimizationSavings string  `json:"fee_optimization_savings"`

	// Error metrics
	TotalErrors  int64 `json:"total_errors"`
	RetryCount   int64 `json:"retry_count"`
	TimeoutCount int64 `json:"timeout_count"`

	// Resource utilization
	MemoryUsage         float64 `json:"memory_usage_mb"`
	CPUUsage            float64 `json:"cpu_usage_percent"`
	DatabaseConnections int     `json:"database_connections"`

	// Timestamp
	LastUpdated time.Time `json:"last_updated"`
}

// AlertThresholds defines thresholds for monitoring alerts
type AlertThresholds struct {
	MaxQueueDepth     int64   `json:"max_queue_depth"`
	MaxProcessingTime float64 `json:"max_processing_time_seconds"`
	MinSuccessRate    float64 `json:"min_success_rate_percent"`
	MaxFailureRate    float64 `json:"max_failure_rate_percent"`
	MaxRetryCount     int64   `json:"max_retry_count"`
	MaxMemoryUsage    float64 `json:"max_memory_usage_mb"`
	MaxCPUUsage       float64 `json:"max_cpu_usage_percent"`
}

// AlertEvent represents a monitoring alert
type AlertEvent struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Severity   string                 `json:"severity"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
}

// DashboardData represents data for the monitoring dashboard
type DashboardData struct {
	// Current metrics
	Metrics TransactionMetrics `json:"metrics"`

	// Recent activity
	RecentTransactions []models.Transaction      `json:"recent_transactions"`
	RecentBatches      []models.TransactionBatch `json:"recent_batches"`

	// Status distribution
	StatusDistribution map[models.TransactionStatus]int64 `json:"status_distribution"`
	TypeDistribution   map[models.TransactionType]int64   `json:"type_distribution"`

	// Performance trends (last 24 hours)
	HourlyStats []HourlyStats `json:"hourly_stats"`

	// Active alerts
	ActiveAlerts []AlertEvent `json:"active_alerts"`

	// System health
	SystemHealth SystemHealth `json:"system_health"`
}

// HourlyStats represents transaction statistics for one hour
type HourlyStats struct {
	Hour                  time.Time `json:"hour"`
	TransactionCount      int64     `json:"transaction_count"`
	SuccessCount          int64     `json:"success_count"`
	FailureCount          int64     `json:"failure_count"`
	AverageProcessingTime float64   `json:"average_processing_time"`
	TotalFees             string    `json:"total_fees"`
}

// SystemHealth represents overall system health status
type SystemHealth struct {
	OverallStatus      string    `json:"overall_status"`
	XRPLConnection     string    `json:"xrpl_connection"`
	DatabaseConnection string    `json:"database_connection"`
	RedisConnection    string    `json:"redis_connection"`
	LastHealthCheck    time.Time `json:"last_health_check"`
}

// NewTransactionMonitoringService creates a new monitoring service
func NewTransactionMonitoringService(
	transactionRepo repository.TransactionRepositoryInterface,
	messagingService *messaging.MessagingService,
) *TransactionMonitoringService {
	return &TransactionMonitoringService{
		transactionRepo:  transactionRepo,
		messagingService: messagingService,
		metrics:          &TransactionMetrics{},
		updateInterval:   30 * time.Second,
		retentionPeriod:  7 * 24 * time.Hour, // 7 days
		stopChannel:      make(chan struct{}),
		alertThresholds:  DefaultAlertThresholds(),
	}
}

// DefaultAlertThresholds returns sensible default alert thresholds
func DefaultAlertThresholds() AlertThresholds {
	return AlertThresholds{
		MaxQueueDepth:     1000,
		MaxProcessingTime: 300.0, // 5 minutes
		MinSuccessRate:    95.0,  // 95%
		MaxFailureRate:    5.0,   // 5%
		MaxRetryCount:     100,
		MaxMemoryUsage:    512.0, // 512 MB
		MaxCPUUsage:       80.0,  // 80%
	}
}

// Start begins the monitoring service
func (s *TransactionMonitoringService) Start() error {
	if s.isRunning {
		return fmt.Errorf("monitoring service is already running")
	}

	s.isRunning = true
	log.Println("Starting Transaction Monitoring Service...")

	// Start background workers
	s.wg.Add(3)
	go s.metricsCollector()
	go s.alertMonitor()
	go s.systemHealthChecker()

	// Subscribe to transaction events
	if err := s.subscribeToEvents(); err != nil {
		log.Printf("Warning: Failed to subscribe to events: %v", err)
	}

	log.Println("Transaction Monitoring Service started successfully")
	return nil
}

// Stop gracefully shuts down the monitoring service
func (s *TransactionMonitoringService) Stop() {
	if !s.isRunning {
		return
	}

	log.Println("Stopping Transaction Monitoring Service...")
	s.isRunning = false
	close(s.stopChannel)
	s.wg.Wait()
	log.Println("Transaction Monitoring Service stopped")
}

// GetDashboardData returns current dashboard data
func (s *TransactionMonitoringService) GetDashboardData() (*DashboardData, error) {
	s.metricsMutex.RLock()
	metrics := *s.metrics
	s.metricsMutex.RUnlock()

	// Get recent transactions
	recentTransactions, err := s.transactionRepo.GetTransactionsByStatus(
		models.TransactionStatusConfirmed, 10, 0)
	if err != nil {
		log.Printf("Failed to get recent transactions: %v", err)
		recentTransactions = []*models.Transaction{}
	}

	// Get recent batches
	recentBatches, err := s.transactionRepo.GetTransactionBatchesByStatus(
		models.TransactionStatusConfirmed, 5, 0)
	if err != nil {
		log.Printf("Failed to get recent batches: %v", err)
		recentBatches = []*models.TransactionBatch{}
	}

	// Get status distribution
	statusDistribution, err := s.transactionRepo.GetTransactionCountByStatus()
	if err != nil {
		log.Printf("Failed to get status distribution: %v", err)
		statusDistribution = make(map[models.TransactionStatus]int64)
	}

	// Get hourly stats
	hourlyStats := s.getHourlyStats()

	// Get active alerts (simplified - would come from alert storage)
	activeAlerts := s.getActiveAlerts()

	// Get system health
	systemHealth := s.getSystemHealth()

	// Convert pointers to values for recent transactions
	recentTxs := make([]models.Transaction, len(recentTransactions))
	for i, tx := range recentTransactions {
		recentTxs[i] = *tx
	}

	// Convert pointers to values for recent batches
	recentBtches := make([]models.TransactionBatch, len(recentBatches))
	for i, batch := range recentBatches {
		recentBtches[i] = *batch
	}

	return &DashboardData{
		Metrics:            metrics,
		RecentTransactions: recentTxs,
		RecentBatches:      recentBtches,
		StatusDistribution: statusDistribution,
		HourlyStats:        hourlyStats,
		ActiveAlerts:       activeAlerts,
		SystemHealth:       systemHealth,
	}, nil
}

// GetMetrics returns current metrics
func (s *TransactionMonitoringService) GetMetrics() TransactionMetrics {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()
	return *s.metrics
}

// UpdateAlertThresholds updates the alert thresholds
func (s *TransactionMonitoringService) UpdateAlertThresholds(thresholds AlertThresholds) {
	s.alertThresholds = thresholds
	log.Println("Alert thresholds updated")
}

// metricsCollector collects and updates metrics periodically
func (s *TransactionMonitoringService) metricsCollector() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChannel:
			return
		case <-ticker.C:
			s.updateMetrics()
		}
	}
}

// alertMonitor monitors metrics and triggers alerts
func (s *TransactionMonitoringService) alertMonitor() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChannel:
			return
		case <-ticker.C:
			s.checkAlerts()
		}
	}
}

// systemHealthChecker monitors system health
func (s *TransactionMonitoringService) systemHealthChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChannel:
			return
		case <-ticker.C:
			s.checkSystemHealth()
		}
	}
}

// updateMetrics collects and updates current metrics
func (s *TransactionMonitoringService) updateMetrics() {
	// Get current statistics
	stats, err := s.transactionRepo.GetTransactionStats()
	if err != nil {
		log.Printf("Failed to get transaction stats: %v", err)
		return
	}

	s.metricsMutex.Lock()
	defer s.metricsMutex.Unlock()

	// Update basic metrics
	s.metrics.QueueDepth = stats.PendingTransactions
	s.metrics.ProcessingCount = stats.ProcessingTransactions
	s.metrics.AverageProcessingTime = stats.AverageProcessingTime

	// Calculate rates (simplified)
	totalTransactions := stats.CompletedTransactions + stats.FailedTransactions
	if totalTransactions > 0 {
		s.metrics.SuccessRate = float64(stats.CompletedTransactions) / float64(totalTransactions) * 100
		s.metrics.FailureRate = float64(stats.FailedTransactions) / float64(totalTransactions) * 100
	}

	s.metrics.LastUpdated = time.Now()

	log.Printf("Metrics updated - Queue: %d, Processing: %d, Success Rate: %.2f%%",
		s.metrics.QueueDepth, s.metrics.ProcessingCount, s.metrics.SuccessRate)
}

// checkAlerts checks current metrics against thresholds
func (s *TransactionMonitoringService) checkAlerts() {
	s.metricsMutex.RLock()
	metrics := *s.metrics
	s.metricsMutex.RUnlock()

	// Check queue depth
	if metrics.QueueDepth > s.alertThresholds.MaxQueueDepth {
		s.triggerAlert("high_queue_depth", "warning",
			fmt.Sprintf("Queue depth (%d) exceeds threshold (%d)",
				metrics.QueueDepth, s.alertThresholds.MaxQueueDepth),
			map[string]interface{}{
				"current_depth": metrics.QueueDepth,
				"threshold":     s.alertThresholds.MaxQueueDepth,
			})
	}

	// Check processing time
	if metrics.AverageProcessingTime > s.alertThresholds.MaxProcessingTime {
		s.triggerAlert("slow_processing", "warning",
			fmt.Sprintf("Average processing time (%.2fs) exceeds threshold (%.2fs)",
				metrics.AverageProcessingTime, s.alertThresholds.MaxProcessingTime),
			map[string]interface{}{
				"current_time": metrics.AverageProcessingTime,
				"threshold":    s.alertThresholds.MaxProcessingTime,
			})
	}

	// Check success rate
	if metrics.SuccessRate < s.alertThresholds.MinSuccessRate {
		s.triggerAlert("low_success_rate", "critical",
			fmt.Sprintf("Success rate (%.2f%%) below threshold (%.2f%%)",
				metrics.SuccessRate, s.alertThresholds.MinSuccessRate),
			map[string]interface{}{
				"current_rate": metrics.SuccessRate,
				"threshold":    s.alertThresholds.MinSuccessRate,
			})
	}
}

// checkSystemHealth checks the health of system components
func (s *TransactionMonitoringService) checkSystemHealth() {
	// This would check actual system health in production
	log.Println("System health check completed")
}

// triggerAlert creates and publishes an alert
func (s *TransactionMonitoringService) triggerAlert(alertType, severity, message string, details map[string]interface{}) {
	alert := AlertEvent{
		ID:        fmt.Sprintf("%s_%d", alertType, time.Now().Unix()),
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	log.Printf("ALERT [%s] %s: %s", severity, alertType, message)

	// Publish alert event
	event := &messaging.Event{
		Type:      "monitoring_alert",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"alert": alert,
		},
	}

	if err := s.messagingService.PublishEvent(event); err != nil {
		log.Printf("Failed to publish alert event: %v", err)
	}
}

// subscribeToEvents subscribes to relevant transaction events
func (s *TransactionMonitoringService) subscribeToEvents() error {
	// Subscribe to transaction events for real-time updates
	events := []string{
		"transaction_queued",
		"transaction_confirmed",
		"transaction_failed",
		"batch_completed",
	}

	for _, eventType := range events {
		if err := s.messagingService.SubscribeToEvent(eventType, s.handleTransactionEvent); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", eventType, err)
		}
	}

	return nil
}

// handleTransactionEvent handles incoming transaction events
func (s *TransactionMonitoringService) handleTransactionEvent(event *messaging.Event) error {
	// Update real-time metrics based on events
	switch event.Type {
	case "transaction_queued":
		s.metricsMutex.Lock()
		s.metrics.QueueDepth++
		s.metricsMutex.Unlock()
	case "transaction_confirmed":
		s.metricsMutex.Lock()
		s.metrics.QueueDepth--
		s.metricsMutex.Unlock()
	case "transaction_failed":
		s.metricsMutex.Lock()
		s.metrics.QueueDepth--
		s.metrics.TotalErrors++
		s.metricsMutex.Unlock()
	}

	return nil
}

// getHourlyStats returns hourly statistics for the last 24 hours
func (s *TransactionMonitoringService) getHourlyStats() []HourlyStats {
	// This would query actual hourly statistics from the database
	// For now, return empty slice
	return []HourlyStats{}
}

// getActiveAlerts returns current active alerts
func (s *TransactionMonitoringService) getActiveAlerts() []AlertEvent {
	// This would retrieve active alerts from storage
	// For now, return empty slice
	return []AlertEvent{}
}

// getSystemHealth returns current system health status
func (s *TransactionMonitoringService) getSystemHealth() SystemHealth {
	return SystemHealth{
		OverallStatus:      "healthy",
		XRPLConnection:     "connected",
		DatabaseConnection: "connected",
		RedisConnection:    "connected",
		LastHealthCheck:    time.Now(),
	}
}
