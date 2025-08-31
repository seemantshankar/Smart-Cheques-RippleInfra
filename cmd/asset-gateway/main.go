package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/smart-payment-infrastructure/internal/middleware"
)

func main() {
	r := gin.New()

	// Add global error handling middleware first
	r.Use(middleware.ErrorHandler())

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "asset-gateway",
			"status":  "healthy",
		})
	})

	log.Println("Asset Gateway starting on :8004")
	log.Fatal(http.ListenAndServe(":8004", r))
}
