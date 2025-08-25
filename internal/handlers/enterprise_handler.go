package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smart-payment-infrastructure/internal/models"
	"github.com/smart-payment-infrastructure/internal/services"
)

// EnterpriseHandler handles enterprise-related HTTP requests
type EnterpriseHandler struct {
	enterpriseService *services.EnterpriseService
}

// NewEnterpriseHandler creates a new enterprise handler
func NewEnterpriseHandler(enterpriseService *services.EnterpriseService) *EnterpriseHandler {
	return &EnterpriseHandler{
		enterpriseService: enterpriseService,
	}
}

// RegisterEnterprise handles enterprise registration
func (h *EnterpriseHandler) RegisterEnterprise(c *gin.Context) {
	var req models.EnterpriseRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	enterprise, err := h.enterpriseService.RegisterEnterprise(&req)
	if err != nil {
		switch err {
		case services.ErrEnterpriseAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error": "Enterprise with this registration number already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to register enterprise",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Enterprise registered successfully",
		"enterprise": enterprise,
	})
}

// GetEnterprise retrieves an enterprise by ID
func (h *EnterpriseHandler) GetEnterprise(c *gin.Context) {
	enterpriseIDStr := c.Param("id")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	enterprise, err := h.enterpriseService.GetEnterpriseByID(enterpriseID)
	if err != nil {
		switch err {
		case services.ErrEnterpriseNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Enterprise not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve enterprise",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enterprise": enterprise,
	})
}

// UpdateKYBStatus updates the KYB status of an enterprise
func (h *EnterpriseHandler) UpdateKYBStatus(c *gin.Context) {
	enterpriseIDStr := c.Param("id")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	var req models.KYBVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	var status models.KYBStatus
	switch req.Action {
	case "approve":
		status = models.KYBStatusVerified
	case "reject":
		status = models.KYBStatusRejected
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid action. Must be 'approve' or 'reject'",
		})
		return
	}

	if err := h.enterpriseService.UpdateKYBStatus(enterpriseID, status, req.Comments); err != nil {
		switch err {
		case services.ErrEnterpriseNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Enterprise not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update KYB status",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KYB status updated successfully",
		"status":  status,
	})
}

// UploadDocument handles document upload for KYB
func (h *EnterpriseHandler) UploadDocument(c *gin.Context) {
	enterpriseIDStr := c.Param("id")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	// Get document type from form
	docTypeStr := c.PostForm("document_type")
	if docTypeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Document type is required",
		})
		return
	}

	docType := models.DocumentType(docTypeStr)

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File is required",
		})
		return
	}
	defer file.Close()

	document, err := h.enterpriseService.UploadDocument(enterpriseID, docType, file, header)
	if err != nil {
		switch err {
		case services.ErrEnterpriseNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Enterprise not found",
			})
		case services.ErrFileTooLarge:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File too large. Maximum size is 10MB",
			})
		case services.ErrInvalidFileType:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid file type. Allowed types: PDF, JPEG, PNG, DOC, DOCX",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to upload document",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Document uploaded successfully",
		"document": document,
	})
}

// GetDocuments retrieves all documents for an enterprise
func (h *EnterpriseHandler) GetDocuments(c *gin.Context) {
	enterpriseIDStr := c.Param("id")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	documents, err := h.enterpriseService.GetEnterpriseDocuments(enterpriseID)
	if err != nil {
		switch err {
		case services.ErrEnterpriseNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Enterprise not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve documents",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
	})
}

// VerifyDocument updates the verification status of a document
func (h *EnterpriseHandler) VerifyDocument(c *gin.Context) {
	docIDStr := c.Param("docId")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid document ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=verified rejected"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	var status models.DocumentStatus
	switch req.Status {
	case "verified":
		status = models.DocumentStatusVerified
	case "rejected":
		status = models.DocumentStatusRejected
	}

	if err := h.enterpriseService.VerifyDocument(docID, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update document status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Document status updated successfully",
		"status":  status,
	})
}

// PerformKYBCheck initiates automated KYB checks
func (h *EnterpriseHandler) PerformKYBCheck(c *gin.Context) {
	enterpriseIDStr := c.Param("id")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	if err := h.enterpriseService.PerformKYBCheck(enterpriseID); err != nil {
		switch err {
		case services.ErrEnterpriseNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Enterprise not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to perform KYB check",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KYB check initiated successfully",
	})
}

// PerformComplianceCheck initiates automated compliance checks
func (h *EnterpriseHandler) PerformComplianceCheck(c *gin.Context) {
	enterpriseIDStr := c.Param("id")
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid enterprise ID",
		})
		return
	}

	if err := h.enterpriseService.PerformComplianceCheck(enterpriseID); err != nil {
		switch err {
		case services.ErrEnterpriseNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Enterprise not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to perform compliance check",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Compliance check initiated successfully",
	})
}