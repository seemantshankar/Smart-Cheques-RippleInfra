package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// OracleHandler handles HTTP requests for oracle functionality
type OracleHandler struct {
	oracleService       *services.OracleService
	verificationService *services.OracleVerificationService
	monitoringService   *services.OracleMonitoringService
}

// NewOracleHandler creates a new oracle handler
func NewOracleHandler(
	oracleService *services.OracleService,
	verificationService *services.OracleVerificationService,
	monitoringService *services.OracleMonitoringService,
) *OracleHandler {
	return &OracleHandler{
		oracleService:       oracleService,
		verificationService: verificationService,
		monitoringService:   monitoringService,
	}
}

// RegisterProvider registers a new oracle provider
func (h *OracleHandler) RegisterProvider(w http.ResponseWriter, r *http.Request) {
	var provider models.OracleProvider
	if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Set default values
	provider.ID = uuid.New()
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()
	provider.IsActive = true
	provider.Reliability = 1.0

	if err := h.oracleService.RegisterProvider(r.Context(), &provider); err != nil {
		http.Error(w, fmt.Sprintf("Failed to register provider: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(provider); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetProvider retrieves an oracle provider by ID
func (h *OracleHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	provider, err := h.oracleService.GetProvider(r.Context(), providerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get provider: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(provider); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// UpdateProvider updates an existing oracle provider
func (h *OracleHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	var provider models.OracleProvider
	if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Set the ID from the URL
	provider.ID = providerID
	provider.UpdatedAt = time.Now()

	if err := h.oracleService.UpdateProvider(r.Context(), &provider); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update provider: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(provider); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// DeleteProvider deletes an oracle provider
func (h *OracleHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	if err := h.oracleService.DeleteProvider(r.Context(), providerID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete provider: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProviders lists all oracle providers
func (h *OracleHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 100
	offset := 0

	providers, err := h.oracleService.ListProviders(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list providers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(providers); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetActiveProviders retrieves all active oracle providers
func (h *OracleHandler) GetActiveProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.oracleService.GetActiveProviders(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get active providers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(providers); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetProvidersByType retrieves oracle providers by type
func (h *OracleHandler) GetProvidersByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerType := models.OracleType(vars["type"])

	providers, err := h.oracleService.GetProvidersByType(r.Context(), providerType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get providers by type: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(providers); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// HealthCheck performs a health check on a provider
func (h *OracleHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	status, err := h.oracleService.HealthCheck(r.Context(), providerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to perform health check: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetRequest retrieves an oracle request by ID
func (h *OracleHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	request, err := h.oracleService.GetRequest(r.Context(), requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get request: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(request); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// VerifyMilestone verifies a milestone using oracle providers
func (h *OracleHandler) VerifyMilestone(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MilestoneID  string               `json:"milestone_id"`
		Condition    string               `json:"condition"`
		OracleConfig *models.OracleConfig `json:"oracle_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.MilestoneID == "" || req.Condition == "" || req.OracleConfig == nil {
		http.Error(w, "Missing required fields: milestone_id, condition, oracle_config", http.StatusBadRequest)
		return
	}

	response, err := h.verificationService.VerifyMilestone(r.Context(), req.MilestoneID, req.Condition, req.OracleConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify milestone: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetVerificationResult retrieves the result of a previous verification
func (h *OracleHandler) GetVerificationResult(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["request_id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	response, err := h.verificationService.GetVerificationResult(r.Context(), requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get verification result: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetProof retrieves verification evidence for a completed verification
func (h *OracleHandler) GetProof(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID, err := uuid.Parse(vars["request_id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	evidence, err := h.verificationService.GetProof(r.Context(), requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get proof: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := w.Write(evidence); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetDashboardMetrics retrieves metrics for the oracle monitoring dashboard
func (h *OracleHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.monitoringService.GetDashboardMetrics(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get dashboard metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetSLAMonitoring retrieves SLA monitoring data for oracle providers
func (h *OracleHandler) GetSLAMonitoring(w http.ResponseWriter, r *http.Request) {
	report, err := h.monitoringService.GetSLAMonitoring(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get SLA monitoring report: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetCostAnalysis retrieves cost analysis for oracle usage
func (h *OracleHandler) GetCostAnalysis(w http.ResponseWriter, r *http.Request) {
	report, err := h.monitoringService.GetCostAnalysis(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get cost analysis report: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(report); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
