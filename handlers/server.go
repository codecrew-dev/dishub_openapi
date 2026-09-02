package handlers

import (
	"net/http"
	"strings"
	"time"

	"dishub_openapi/database"
	"dishub_openapi/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetServerList(c *gin.Context) {
	query := c.DefaultQuery("query", c.DefaultQuery("q", ""))
	tagsStr := c.Query("tags")
	p := parsePagination(c)

	collection := database.GetCollection("servers")
	filter := bson.M{}
	if query != "" {
		filter["name"] = bson.M{"$regex": query, "$options": "i"}
	}
	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		filter["tags"] = bson.M{"$in": tags}
	}

	total, err := collection.CountDocuments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count servers"})
		return
	}

	findOpts := options.Find().SetLimit(p.Limit).SetSkip(p.Skip)
	cursor, err := collection.Find(c.Request.Context(), filter, findOpts)
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
	if servers == nil {
		servers = []models.ServerResponse{}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse[models.ServerResponse]{
		Data:       servers,
		Total:      total,
		Page:       p.Page,
		TotalPages: calcTotalPages(total, p.Limit),
		Limit:      p.Limit,
	})
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

	activityCollection := database.GetCollection("serveractivities")
	activityFilter := bson.M{
		"serverId": serverID,
		"date":     bson.M{"$gte": time.Now().Add(-30 * 24 * time.Hour)},
	}

	findOptions := options.Find().SetSort(bson.M{"date": 1})
	cursorAct, actErr := activityCollection.Find(c.Request.Context(), activityFilter, findOptions)
	if actErr == nil {
		var activities []models.ServerActivityResponse
		if actDecodeErr := cursorAct.All(c.Request.Context(), &activities); actDecodeErr == nil {
			if len(activities) > 0 {
				server.Activities = activities
			} else {
				server.Activities = []models.ServerActivityResponse{}
			}
		}
		cursorAct.Close(c.Request.Context())
	} else {
		server.Activities = []models.ServerActivityResponse{}
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
