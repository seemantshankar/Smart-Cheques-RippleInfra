package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
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