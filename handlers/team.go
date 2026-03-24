package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"dishub_openapi/database"
	"dishub_openapi/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func expandAvatar(avatar string) string {
	if avatar == "" || strings.HasPrefix(avatar, "http") {
		return avatar
	}
	parts := strings.Split(avatar, "/")
	if len(parts) == 3 {
		t := parts[0]
		id := parts[1]
		hash := parts[2]
		ext := "webp"
		if strings.HasPrefix(hash, "a_") {
			ext = "gif"
		}
		return fmt.Sprintf("https://cdn.discordapp.com/%s/%s/%s.%s?size=1024", t, id, hash, ext)
	}
	return avatar
}

func GetTeamList(c *gin.Context) {
	query := c.DefaultQuery("query", "")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	collection := database.GetCollection("teams")
	filter := bson.M{}
	if query != "" {
		filter["name"] = bson.M{"$regex": query, "$options": "i"}
	}

	cursor, err := collection.Find(c.Request.Context(), filter, options.Find().SetLimit(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var teams []models.TeamResponse
	if err = cursor.All(c.Request.Context(), &teams); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode teams"})
		return
	}

	for i := range teams {
		teams[i].Avatar = expandAvatar(teams[i].Avatar)
	}

	c.JSON(http.StatusOK, teams)
}

func GetTeamInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	collection := database.GetCollection("teams")
	var team models.TeamResponse
	err = collection.FindOne(c.Request.Context(), bson.M{"_id": id}).Decode(&team)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	team.Avatar = expandAvatar(team.Avatar)

	c.JSON(http.StatusOK, team)
}
