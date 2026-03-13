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

func GetBotList(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	collection := database.GetCollection("bots")
	filter := bson.M{}
	if query != "" {
		filter["name"] = bson.M{"$regex": query, "$options": "i"}
	}

	cursor, err := collection.Find(c.Request.Context(), filter, options.Find().SetLimit(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bots"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var bots []models.BotResponse
	if err = cursor.All(c.Request.Context(), &bots); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode bots"})
		return
	}

	for i := range bots {
		bots[i].Library = bots[i].Submission.Library
		bots[i].Website = bots[i].Submission.Website
		bots[i].SupportServer = bots[i].Submission.SupportServer
		bots[i].InviteUrl = bots[i].Submission.InviteUrl
		bots[i].BotLangs = bots[i].Submission.BotLangs
		bots[i].ShortDescs = bots[i].Submission.ShortDescs
		bots[i].LongDescs = bots[i].Submission.LongDescs
	}

	c.JSON(http.StatusOK, bots)
}

func GetBotInfo(c *gin.Context) {
	botID := c.Param("id")
	collection := database.GetCollection("bots")

	var bot models.BotResponse
	err := collection.FindOne(c.Request.Context(), bson.M{"botId": botID}).Decode(&bot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bot not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	bot.Library = bot.Submission.Library
	bot.Website = bot.Submission.Website
	bot.SupportServer = bot.Submission.SupportServer
	bot.InviteUrl = bot.Submission.InviteUrl
	bot.BotLangs = bot.Submission.BotLangs
	bot.ShortDescs = bot.Submission.ShortDescs
	bot.LongDescs = bot.Submission.LongDescs

	c.JSON(http.StatusOK, bot)
}

func CheckBotVote(c *gin.Context) {
	botID := c.Param("id")
	userID := c.Query("userID")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID query parameter required"})
		return
	}

	// Verify if the token belongs to this bot
	appVal, _ := c.Get("app")
	app := appVal.(models.DeveloperApp)
	if app.TargetType != "bot" || app.TargetId != botID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this bot"})
		return
	}

	collection := database.GetCollection("votes")
	
	// Check if vote exists and was updated within the last 12 hours
	count, err := collection.CountDocuments(c.Request.Context(), bson.M{
		"botId":     botID,
		"userId":    userID,
		"updatedAt": bson.M{"$gte": time.Now().Add(-12 * time.Hour)},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, models.BotVotedResponse{Voted: count > 0})
}

func UpdateBotStats(c *gin.Context) {
	botID := c.Param("id")
	var req models.BotStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify if the token belongs to this bot
	appVal, _ := c.Get("app")
	app := appVal.(models.DeveloperApp)
	if app.TargetType != "bot" || app.TargetId != botID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this bot"})
		return
	}

	collection := database.GetCollection("bots")
	_, err := collection.UpdateOne(
		c.Request.Context(),
		bson.M{"botId": botID},
		bson.M{"$set": bson.M{
			"serverCount": req.Servers,
			"shards":      req.Shards,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "servers": req.Servers, "shards": req.Shards, "message": "Stats updated successfully"})
}
