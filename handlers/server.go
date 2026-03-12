package handlers

import (
	"net/http"
	"strconv"
	"time"

	"dishub_openapi/database"
	"dishub_openapi/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetServerList(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	collection := database.GetCollection("servers")
	filter := bson.M{}
	if query != "" {
		filter["name"] = bson.M{"$regex": query, "$options": "i"}
	}

	cursor, err := collection.Find(c.Request.Context(), filter, options.Find().SetLimit(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var servers []models.ServerResponse
	if err = cursor.All(c.Request.Context(), &servers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode servers"})
		return
	}

	c.JSON(http.StatusOK, servers)
}

func GetServerInfo(c *gin.Context) {
	serverID := c.Param("id")
	collection := database.GetCollection("servers")

	var server models.ServerResponse
	err := collection.FindOne(c.Request.Context(), bson.M{"serverId": serverID}).Decode(&server)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, server)
}

func CheckServerVote(c *gin.Context) {
	serverID := c.Param("id")
	userID := c.Query("userID")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID query parameter required"})
		return
	}

	// Verify if the token belongs to this server
	appVal, _ := c.Get("app")
	app := appVal.(models.DeveloperApp)
	if app.TargetType != "server" || app.TargetId != serverID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this server"})
		return
	}

	collection := database.GetCollection("servervotes")
	
	// Check if vote exists and was updated within the last 12 hours
	count, err := collection.CountDocuments(c.Request.Context(), bson.M{
		"serverId":  serverID,
		"userId":    userID,
		"updatedAt": bson.M{"$gte": time.Now().Add(-12 * time.Hour)},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, models.ServerVotedResponse{Voted: count > 0})
}
