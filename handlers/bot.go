package handlers

import (
	"log"
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

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func GetBotList(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	collection := database.GetCollection("bots")
	filter := bson.M{"verified": true}
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
		// Also handle Banner which is directly mapped from BSON
		if bots[i].Banner != nil && *bots[i].Banner == "" {
			bots[i].Banner = nil
		}
	}

	c.JSON(http.StatusOK, bots)
}

func GetBotInfo(c *gin.Context) {
	botID := c.Param("id")
	collection := database.GetCollection("bots")

	var bot models.BotResponse
	err := collection.FindOne(c.Request.Context(), bson.M{"botId": botID, "verified": true}).Decode(&bot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bot not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Also handle Banner
	if bot.Banner != nil && *bot.Banner == "" {
		bot.Banner = nil
	}

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
	
	// Also check if bot is verified
	botCollection := database.GetCollection("bots")
	var bot models.BotResponse
	err = botCollection.FindOne(c.Request.Context(), bson.M{"botId": botID, "verified": true}).Decode(&bot)
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
	
	// Fetch old stats for webhook
	var bot models.BotResponse
	err := collection.FindOne(c.Request.Context(), bson.M{"botId": botID, "verified": true}).Decode(&bot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bot is not verified or not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	_, err = collection.UpdateOne(
		c.Request.Context(),
		bson.M{"botId": botID, "verified": true},
		bson.M{"$set": bson.M{
			"serverCount": req.Servers,
			"shards":      req.Shards,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stats"})
		return
	}

	// Trigger Webhook
	log.Printf("[Webhook] Triggering for bot %s, app %v", botID, app.ID)
	go func() {
		payload := gin.H{
			"type": "bot",
			"data": gin.H{
				"type":   1, // Server count update
				"botId":  botID,
				"before": bot.Servers,
				"after":  req.Servers,
			},
			"timestamp": time.Now().UnixMilli(),
		}
		SendWebhookNotification(app, payload)

		// Trigger Discord Webhook (Embed)
		hasBotServerCount := false
		for _, e := range app.WebhookEvents {
			if e == "bot.server_count" {
				hasBotServerCount = true
				break
			}
		}

		if app.DiscordWebhookURL != "" && hasBotServerCount {
			log.Printf("[Webhook] Sending Discord embed to %s", app.DiscordWebhookURL)
			embed := models.DiscordEmbed{
				Author: &models.DiscordAuthor{
					Name:    bot.Name,
					IconURL: bot.Avatar,
					URL:     "https://dishub.codecrew.kr/bots/" + botID,
				},
				Title: "📊 서버 수 변동",
				Color: 0x5865F2,
				Fields: []models.DiscordEmbedField{
					{Name: "이전", Value: "`" + strconv.Itoa(bot.Servers) + "`개", Inline: true},
					{Name: "이후", Value: "`" + strconv.Itoa(req.Servers) + "`개", Inline: true},
				},
				Timestamp: time.Now().Format(time.RFC3339),
				Footer:    &models.DiscordFooter{Text: "DisHub"},
			}
			SendDiscordWebhookEmbed(app.DiscordWebhookURL, embed)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "servers": req.Servers, "shards": req.Shards, "message": "Stats updated successfully"})
}
