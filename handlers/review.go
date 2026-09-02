package handlers

import (
	"context"
	"net/http"
	"time"

	"dishub_openapi/database"
	"dishub_openapi/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// recalcRating recalculates the average rating and review count for a target and updates it.
func recalcRating(ctx context.Context, targetType, targetId string) {
	reviewCollection := database.GetCollection("reviews")

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"targetType": targetType, "targetId": targetId}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "avg", Value: bson.M{"$avg": "$rating"}},
			{Key: "count", Value: bson.M{"$sum": 1}},
		}}},
	}

	cursor, err := reviewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	var results []struct {
		Avg   float64 `bson:"avg"`
		Count int     `bson:"count"`
	}
	cursor.All(ctx, &results)

	var avg float64
	var count int
	if len(results) > 0 {
		avg = results[0].Avg
		count = results[0].Count
	}

	var collection *mongo.Collection
	var filter bson.M
	if targetType == "bot" {
		collection = database.GetCollection("bots")
		filter = bson.M{"botId": targetId}
	} else {
		collection = database.GetCollection("servers")
		filter = bson.M{"serverId": targetId}
	}

	collection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"rating": avg, "reviewCount": count}})
}

// GetReviewsHandler returns a handler that lists reviews for a bot or server.
func GetReviewsHandler(targetType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetId := c.Param("id")
		p := parsePagination(c)

		appVal, _ := c.Get("app")
		app := appVal.(models.DeveloperApp)
		if app.TargetType != targetType || app.TargetId != targetId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this " + targetType})
			return
		}

		collection := database.GetCollection("reviews")
		filter := bson.M{"targetType": targetType, "targetId": targetId}

		total, err := collection.CountDocuments(c.Request.Context(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count reviews"})
			return
		}

		findOpts := options.Find().
			SetLimit(p.Limit).
			SetSkip(p.Skip).
			SetSort(bson.M{"createdAt": -1})

		cursor, err := collection.Find(c.Request.Context(), filter, findOpts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
			return
		}
		defer cursor.Close(c.Request.Context())

		var reviews []models.ReviewModel
		if err = cursor.All(c.Request.Context(), &reviews); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode reviews"})
			return
		}
		if reviews == nil {
			reviews = []models.ReviewModel{}
		}

		c.JSON(http.StatusOK, models.PaginatedResponse[models.ReviewModel]{
			Data:       reviews,
			Total:      total,
			Page:       p.Page,
			TotalPages: calcTotalPages(total, p.Limit),
			Limit:      p.Limit,
		})
	}
}

// CreateReviewHandler returns a handler that creates a review for a bot or server.
func CreateReviewHandler(targetType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetId := c.Param("id")

		appVal, _ := c.Get("app")
		app := appVal.(models.DeveloperApp)
		if app.TargetType != targetType || app.TargetId != targetId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this " + targetType})
			return
		}

		var req models.ReviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: userId, rating(1-5), content are required"})
			return
		}

		collection := database.GetCollection("reviews")

		// One review per user per target
		existing := collection.FindOne(c.Request.Context(), bson.M{
			"targetType": targetType,
			"targetId":   targetId,
			"userId":     req.UserID,
		})
		if existing.Err() == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already has a review for this " + targetType})
			return
		}

		now := time.Now()
		review := models.ReviewModel{
			ID:         primitive.NewObjectID(),
			TargetType: targetType,
			TargetId:   targetId,
			UserID:     req.UserID,
			Rating:     req.Rating,
			Content:    req.Content,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		_, err := collection.InsertOne(c.Request.Context(), review)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
			return
		}

		recalcRating(c.Request.Context(), targetType, targetId)

		c.JSON(http.StatusCreated, review)
	}
}

// UpdateReviewHandler returns a handler that updates a review.
// Only the original author (userId in body must match) can update.
func UpdateReviewHandler(targetType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetId := c.Param("id")
		reviewIdStr := c.Param("reviewId")

		appVal, _ := c.Get("app")
		app := appVal.(models.DeveloperApp)
		if app.TargetType != targetType || app.TargetId != targetId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this " + targetType})
			return
		}

		reviewId, err := primitive.ObjectIDFromHex(reviewIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return
		}

		var req models.ReviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: userId, rating(1-5), content are required"})
			return
		}

		collection := database.GetCollection("reviews")

		// Verify the review belongs to this user
		var existing models.ReviewModel
		if err := collection.FindOne(c.Request.Context(), bson.M{
			"_id":        reviewId,
			"targetType": targetType,
			"targetId":   targetId,
		}).Decode(&existing); err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			}
			return
		}

		if existing.UserID != req.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot edit another user's review"})
			return
		}

		now := time.Now()
		_, err = collection.UpdateOne(
			c.Request.Context(),
			bson.M{"_id": reviewId},
			bson.M{"$set": bson.M{
				"rating":    req.Rating,
				"content":   req.Content,
				"updatedAt": now,
			}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
			return
		}

		recalcRating(c.Request.Context(), targetType, targetId)

		existing.Rating = req.Rating
		existing.Content = req.Content
		existing.UpdatedAt = now
		c.JSON(http.StatusOK, existing)
	}
}

// DeleteReviewHandler returns a handler that deletes a review.
// Bot/server owner (via token) can delete any review (moderation).
func DeleteReviewHandler(targetType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetId := c.Param("id")
		reviewIdStr := c.Param("reviewId")

		appVal, _ := c.Get("app")
		app := appVal.(models.DeveloperApp)
		if app.TargetType != targetType || app.TargetId != targetId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Token mismatch for this " + targetType})
			return
		}

		reviewId, err := primitive.ObjectIDFromHex(reviewIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return
		}

		collection := database.GetCollection("reviews")
		result, err := collection.DeleteOne(c.Request.Context(), bson.M{
			"_id":        reviewId,
			"targetType": targetType,
			"targetId":   targetId,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
			return
		}
		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			return
		}

		recalcRating(c.Request.Context(), targetType, targetId)

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review deleted successfully"})
	}
}
