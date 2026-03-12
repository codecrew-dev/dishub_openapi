package main

import (
	"log"

	"os"
	"dishub_openapi/database"
	"dishub_openapi/handlers"
	"dishub_openapi/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// DB Connection
	dbURI := os.Getenv("MONGODB_URI")
	if dbURI == "" {
		log.Fatal("MONGODB_URI environment variable is required")
	}
	database.ConnectDB(dbURI)

	r := gin.Default()

	// CORS or other middleware could go here
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.RateLimitMiddleware())

	// Bot routes
	bots := r.Group("/bots")
	{
		// Auth required endpoints
		authorized := bots.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			authorized.GET("", handlers.GetBotList)
			authorized.GET("/:id", handlers.GetBotInfo)
			authorized.GET("/:id/voted", handlers.CheckBotVote)
			authorized.POST("/:id/stats", handlers.UpdateBotStats)
		}
	}

	// Server routes
	servers := r.Group("/servers")
	{
		authorized := servers.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			authorized.GET("", handlers.GetServerList)
			authorized.GET("/:id", handlers.GetServerInfo)
			authorized.GET("/:id/voted", handlers.CheckServerVote)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}
	log.Printf("Server starting on port %s", port)
	r.Run("0.0.0.0:" + port)
}
