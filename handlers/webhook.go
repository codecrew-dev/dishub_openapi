package handlers

import (
	"encoding/json"
	"net/http"
	"dishub_openapi/database"
	"dishub_openapi/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
    "bytes"
    "time"
    "errors"
)

type WebhookUpdateRequest struct {
	URL    string   `json:"url"`
	Secret string   `json:"secret"`
	Events []string `json:"events"`
}

func UpdateWebhook(c *gin.Context) {
	app, exists := c.Get("app")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "App context not found"})
		return
	}
	devApp := app.(models.DeveloperApp)

	var req WebhookUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update in DB
	collection := database.GetCollection("developerapps")
	_, err := collection.UpdateOne(c.Request.Context(), bson.M{"_id": devApp.ID}, bson.M{
		"$set": bson.M{
			"webhookURL":    req.URL,
			"webhookSecret": req.Secret,
			"webhookEvents": req.Events,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Webhook updated successfully"})
}

func VerifyWebhook(c *gin.Context) {
	var req struct {
		URL    string `json:"url"`
		Secret string `json:"secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Send GET request to developer's URL with secret query param
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(req.URL + "?secret=" + req.Secret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to reach webhook URL: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook server returned non-200 status"})
		return
	}

	var verifResp struct {
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&verifResp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode webhook response"})
		return
	}

	if verifResp.Secret != req.Secret {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Secret mismatch in webhook response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Webhook verified successfully"})
}

// SendWebhook logic (placeholder, should be in a utility for reuse)
func SendWebhookNotification(devApp models.DeveloperApp, payload interface{}) error {
    if devApp.WebhookURL == "" {
        return nil
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", devApp.WebhookURL, bytes.NewBuffer(body))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-DisHub-Signature", devApp.WebhookSecret)

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return errors.New("webhook returned status: " + resp.Status)
    }

    return nil
}
