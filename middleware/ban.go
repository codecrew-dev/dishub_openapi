package middleware

import (
	"net/http"

	"dishub_openapi/database"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func IpBanMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		collection := database.GetCollection("openapibans")
		err := collection.FindOne(c.Request.Context(), bson.M{"targetId": clientIP, "type": "ip"}).Decode(&bson.M{})
		
		if err == nil {
			// Found a ban record
			c.JSON(http.StatusForbidden, gin.H{"error": "This IP is banned from accessing the API"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
