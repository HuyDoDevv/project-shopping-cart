package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
)


func ApiKeyMiddleware() gin.HandlerFunc {
	expectedKey := os.Getenv("API_KEY")
	println("Expected API Key:", expectedKey)
	if expectedKey == "" {
		expectedKey = "secret-key" // Default value for local development
	}

	return func(ctx *gin.Context) {
		apiKey := ctx.GetHeader("X-API-KEY")
		if apiKey == "" {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "API Key is required"})
			return
		}

		if apiKey != expectedKey {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		ctx.Next()
	}
}
