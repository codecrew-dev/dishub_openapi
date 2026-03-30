package handlers

import (
	"net/http"

	"strconv"
	"dishub_openapi/database"
	"dishub_openapi/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetUserList(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	collection := database.GetCollection("users")
	filter := bson.M{}
	if query != "" {
		filter["$or"] = []bson.M{
			{"username": bson.M{"$regex": query, "$options": "i"}},
			{"globalName": bson.M{"$regex": query, "$options": "i"}},
		}
	}

	cursor, err := collection.Find(c.Request.Context(), filter, options.Find().SetLimit(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var users []models.UserResponse
	if err = cursor.All(c.Request.Context(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func GetUserInfo(c *gin.Context) {
	userID := c.Param("id")

	// 1. Fetch User info
	userCollection := database.GetCollection("users")
	var user models.UserResponse
	err := userCollection.FindOne(c.Request.Context(), bson.M{"discordId": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// 2. Fetch Bots owned by the user
	botCollection := database.GetCollection("bots")
	botCursor, err := botCollection.Find(c.Request.Context(), bson.M{"ownerId": userID})
	if err == nil {
		var bots []models.BotResponse
		if err = botCursor.All(c.Request.Context(), &bots); err == nil {
			user.Bots = bots
		}
		botCursor.Close(c.Request.Context())
	} else {
		user.Bots = []models.BotResponse{}
	}

	// 3. Fetch Servers owned by the user
	serverCollection := database.GetCollection("servers")
	serverCursor, err := serverCollection.Find(c.Request.Context(), bson.M{"ownerId": userID})
	if err == nil {
		var servers []models.ServerResponse
		if err = serverCursor.All(c.Request.Context(), &servers); err == nil {
			user.Servers = servers
		}
		serverCursor.Close(c.Request.Context())
	} else {
		user.Servers = []models.ServerResponse{}
	}

	c.JSON(http.StatusOK, user)
}
