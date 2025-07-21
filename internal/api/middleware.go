package api

import (
	"evolution-postgres-backup/internal/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := os.Getenv("API_KEY")
		if apiKey == "" {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "API key not configured",
			})
			c.Abort()
			return
		}

		requestApiKey := c.GetHeader("api-key")
		if requestApiKey == "" {
			// Try to get API key from query parameter (for EventSource)
			requestApiKey = c.Query("api-key")
		}

		if requestApiKey == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "api-key header or query parameter required",
			})
			c.Abort()
			return
		}

		if requestApiKey != apiKey {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, api-key, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
