package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"dishub_openapi/database"
	"dishub_openapi/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Token format: dhub_...
		if !strings.HasPrefix(authHeader, "dhub_") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Hash the token
		hash := sha256.Sum256([]byte(authHeader))
		hashedToken := hex.EncodeToString(hash[:])

		// Check if token is banned
		banCollection := database.GetCollection("openapibans")
		err := banCollection.FindOne(c.Request.Context(), bson.M{"targetId": hashedToken, "type": "token"}).Decode(&bson.M{})
		if err == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "This token is banned from accessing the API"})
			c.Abort()
			return
		}

		// Find the app in DB
		collection := database.GetCollection("developerapps")
		var app models.DeveloperApp
		err = collection.FindOne(c.Request.Context(), bson.M{"tokenHash": hashedToken}).Decode(&app)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set app info in context
		c.Set("app", app)
		c.Next()
	}
}
