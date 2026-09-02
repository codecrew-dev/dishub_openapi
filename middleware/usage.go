package middleware

import (
	"time"

	"dishub_openapi/database"
	"dishub_openapi/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// ApiUsageMiddleware records one compact usage event for every authenticated
// OpenAPI request. Logging failures never fail the user's API request.
func ApiUsageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		appValue, exists := c.Get("app")
		app, ok := appValue.(models.DeveloperApp)
		if !exists || !ok {
			return
		}

		_, _ = database.GetCollection("apiusages").InsertOne(c.Request.Context(), bson.M{
			"appId":      app.ID,
			"ownerId":    app.OwnerID,
			"targetType": app.TargetType,
			"targetId":   app.TargetId,
			"method":     c.Request.Method,
			"path":       c.FullPath(),
			"status":     c.Writer.Status(),
			"durationMs": time.Since(startedAt).Milliseconds(),
			"createdAt":  time.Now().UTC(),
		})
	}
}
