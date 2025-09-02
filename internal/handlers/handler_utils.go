package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaginationParams represents parsed pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}

// ParseEnterpriseIDAndPagination parses enterprise ID from URL parameter and pagination parameters from query
func ParseEnterpriseIDAndPagination(c *gin.Context, paramName string) (uuid.UUID, *PaginationParams, error) {
	// Parse enterprise ID
	enterpriseIDStr := c.Param(paramName)
	enterpriseID, err := uuid.Parse(enterpriseIDStr)
	if err != nil {
		return uuid.Nil, nil, err
	}

	// Parse pagination parameters
	params := &PaginationParams{}
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}
	params.Limit = limit

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	params.Offset = offset

	return enterpriseID, params, nil
}

// ParsePaginationParams parses pagination parameters from query string
func ParsePaginationParams(c *gin.Context) *PaginationParams {
	params := &PaginationParams{}
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}
	params.Limit = limit

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	params.Offset = offset

	return params
}

// CreatePaginationResponse creates a standardized pagination response
func CreatePaginationResponse(data interface{}, limit, offset int) gin.H {
	return gin.H{
		"data": data,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  0, // This will be set by the caller if needed
		},
	}
}

// ParseOptionalEnterpriseIDAndPagination parses optional enterprise ID from query and pagination parameters
func ParseOptionalEnterpriseIDAndPagination(c *gin.Context, queryParam string) (*uuid.UUID, *PaginationParams, error) {
	// Parse optional enterprise ID
	var enterpriseID *uuid.UUID
	enterpriseIDStr := c.Query(queryParam)
	if enterpriseIDStr != "" {
		id, err := uuid.Parse(enterpriseIDStr)
		if err != nil {
			return nil, nil, err
		}
		enterpriseID = &id
	}

	// Parse pagination parameters
	params := ParsePaginationParams(c)

	return enterpriseID, params, nil
}

// HandleEnterpriseIDError handles enterprise ID parsing errors
func HandleEnterpriseIDError(c *gin.Context, _ error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid enterprise ID"})
}
