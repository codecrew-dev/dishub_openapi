package handlers

import (
	"bytes"
	"dishub_openapi/database"
	"dishub_openapi/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func applyBotStatsFreshness(bot *models.BotResponse) {
	if bot.LastStatsAt == nil || time.Since(*bot.LastStatsAt) > 24*time.Hour {
		bot.Status = "offline"
	}
}

func GetBotList(c *gin.Context) {
	query := c.DefaultQuery("query", c.DefaultQuery("q", ""))
	tagsStr := c.Query("tags")
	p := parsePagination(c)

	collection := database.GetCollection("bots")
	filter := bson.M{"verified": true}
	if query != "" {
		filter["name"] = bson.M{"$regex": query, "$options": "i"}
	}
	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		filter["tags"] = bson.M{"$in": tags}
	}

	total, err := collection.CountDocuments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count bots"})
		return
	}

	findOpts := options.Find().SetLimit(p.Limit).SetSkip(p.Skip)
	cursor, err := collection.Find(c.Request.Context(), filter, findOpts)
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
	if bots == nil {
		bots = []models.BotResponse{}
	}

	for i := range bots {
		if bots[i].Banner != nil && *bots[i].Banner == "" {
			bots[i].Banner = nil
		}
		applyBotStatsFreshness(&bots[i])
	}

	c.JSON(http.StatusOK, models.PaginatedResponse[models.BotResponse]{
		Data:       bots,
		Total:      total,
		Page:       p.Page,
		TotalPages: calcTotalPages(total, p.Limit),
		Limit:      p.Limit,
	})
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
	applyBotStatsFreshness(&bot)

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
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = "online"
	}
	if status != "online" && status != "idle" && status != "dnd" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Allowed values: online, idle, dnd"})
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

	now := time.Now()
	updateFields := bson.M{
		"serverCount": req.Servers,
		"shards":      req.Shards,
		"status":      status,
		"online":      true,
		"lastStatsAt": now,
	}
	_, err = collection.UpdateOne(
		c.Request.Context(),
		bson.M{"botId": botID, "verified": true},
		bson.M{"$set": updateFields},
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

		// System Log (Discord Audit) - Bot IPC Stats
		go func() {
			ipcURL := os.Getenv("BOT_IPC_URL")
			if ipcURL != "" {
				ipcURL = strings.TrimRight(ipcURL, "/") + "/stats_log"
				statsPayload := map[string]interface{}{
					"botId":       botID,
					"serverCount": req.Servers,
					"shards":      req.Shards,
					"status":      status,
				}
				jsonStr, _ := json.Marshal(statsPayload)
				r, err := http.NewRequest("POST", ipcURL, bytes.NewBuffer(jsonStr))
				if err == nil {
					r.Header.Set("Authorization", os.Getenv("BOT_SYNC_TOKEN"))
					r.Header.Set("Content-Type", "application/json")
					client := &http.Client{}
					client.Do(r)
				}
			}
		}()

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

	c.JSON(http.StatusOK, gin.H{"success": true, "servers": req.Servers, "shards": req.Shards, "status": status, "message": "Stats updated successfully"})
}
