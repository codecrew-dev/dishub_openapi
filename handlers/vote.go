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

type VoteRequest struct {
	UserID string `json:"userId" binding:"required"`
}

func VoteBot(c *gin.Context) {
	botID := c.Param("id")

	var req VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	appVal, _ := c.Get("app")
	app := appVal.(models.DeveloperApp)
	if app.TargetType != "bot" || app.TargetId != botID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this bot"})
		return
	}

	botCollection := database.GetCollection("bots")
	var bot models.BotResponse
	if err := botCollection.FindOne(c.Request.Context(), bson.M{"botId": botID, "verified": true}).Decode(&bot); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bot not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	voteCollection := database.GetCollection("votes")
	count, err := voteCollection.CountDocuments(c.Request.Context(), bson.M{
		"botId":     botID,
		"userId":    req.UserID,
		"updatedAt": bson.M{"$gte": time.Now().Add(-12 * time.Hour)},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Already voted within the last 12 hours"})
		return
	}

	now := time.Now()
	_, err = voteCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"botId": botID, "userId": req.UserID},
		bson.M{
			"$set":         bson.M{"updatedAt": now},
			"$setOnInsert": bson.M{"createdAt": now, "botId": botID, "userId": req.UserID},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	_, err = botCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"botId": botID},
		bson.M{"$inc": bson.M{"hearts": 1}},
	)
	if err != nil {
		log.Printf("[Vote] Failed to increment hearts for bot %s: %v", botID, err)
	}

	newVotes := bot.Votes + 1
	go func() {
		payload := gin.H{
			"type": "bot",
			"data": gin.H{
				"type":   2,
				"botId":  botID,
				"userId": req.UserID,
				"votes":  newVotes,
			},
			"timestamp": now.UnixMilli(),
		}
		SendWebhookNotification(app, payload)

		hasVoteEvent := false
		for _, e := range app.WebhookEvents {
			if e == "bot.vote" {
				hasVoteEvent = true
				break
			}
		}
		if app.DiscordWebhookURL != "" && hasVoteEvent {
			embed := models.DiscordEmbed{
				Author: &models.DiscordAuthor{
					Name:    bot.Name,
					IconURL: bot.Avatar,
					URL:     "https://dishub.codecrew.kr/bots/" + botID,
				},
				Title: "❤️ 새로운 추천",
				Color: 0xFF0000,
				Fields: []models.DiscordEmbedField{
					{Name: "유저 ID", Value: req.UserID, Inline: true},
					{Name: "총 추천 수", Value: "`" + strconv.Itoa(newVotes) + "`개", Inline: true},
				},
				Timestamp: now.Format(time.RFC3339),
				Footer:    &models.DiscordFooter{Text: "DisHub"},
			}
			SendDiscordWebhookEmbed(app.DiscordWebhookURL, embed)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Vote recorded successfully"})
}

func VoteServer(c *gin.Context) {
	serverID := c.Param("id")

	var req VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	appVal, _ := c.Get("app")
	app := appVal.(models.DeveloperApp)
	if app.TargetType != "server" || app.TargetId != serverID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this server"})
		return
	}

	serverCollection := database.GetCollection("servers")
	var server models.ServerResponse
	if err := serverCollection.FindOne(c.Request.Context(), bson.M{"serverId": serverID}).Decode(&server); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	voteCollection := database.GetCollection("servervotes")
	count, err := voteCollection.CountDocuments(c.Request.Context(), bson.M{
		"serverId":  serverID,
		"userId":    req.UserID,
		"updatedAt": bson.M{"$gte": time.Now().Add(-12 * time.Hour)},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Already voted within the last 12 hours"})
		return
	}

	now := time.Now()
	_, err = voteCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"serverId": serverID, "userId": req.UserID},
		bson.M{
			"$set":         bson.M{"updatedAt": now},
			"$setOnInsert": bson.M{"createdAt": now, "serverId": serverID, "userId": req.UserID},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	_, err = serverCollection.UpdateOne(
		c.Request.Context(),
		bson.M{"serverId": serverID},
		bson.M{"$inc": bson.M{"hearts": 1}},
	)
	if err != nil {
		log.Printf("[Vote] Failed to increment hearts for server %s: %v", serverID, err)
	}

	newVotes := server.Votes + 1
	go func() {
		payload := gin.H{
			"type": "server",
			"data": gin.H{
				"type":     2,
				"serverId": serverID,
				"userId":   req.UserID,
				"votes":    newVotes,
			},
			"timestamp": now.UnixMilli(),
		}
		SendWebhookNotification(app, payload)

		hasVoteEvent := false
		for _, e := range app.WebhookEvents {
			if e == "server.vote" {
				hasVoteEvent = true
				break
			}
		}
		if app.DiscordWebhookURL != "" && hasVoteEvent {
			embed := models.DiscordEmbed{
				Author: &models.DiscordAuthor{
					Name:    server.Name,
					IconURL: server.Icon,
					URL:     "https://dishub.codecrew.kr/servers/" + serverID,
				},
				Title: "❤️ 새로운 추천",
				Color: 0xFF0000,
				Fields: []models.DiscordEmbedField{
					{Name: "유저 ID", Value: req.UserID, Inline: true},
					{Name: "총 추천 수", Value: "`" + strconv.Itoa(newVotes) + "`개", Inline: true},
				},
				Timestamp: now.Format(time.RFC3339),
				Footer:    &models.DiscordFooter{Text: "DisHub"},
			}
			SendDiscordWebhookEmbed(app.DiscordWebhookURL, embed)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Vote recorded successfully"})
}
