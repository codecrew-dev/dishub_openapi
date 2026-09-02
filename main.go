package main

import (
	"dishub_openapi/database"
	"dishub_openapi/handlers"
	"dishub_openapi/middleware"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

	r := gin.New()
	var trustedProxies []string
	if configured := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES")); configured != "" {
		trustedProxies = strings.Split(configured, ",")
		for i := range trustedProxies {
			trustedProxies[i] = strings.TrimSpace(trustedProxies[i])
		}
	}
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		log.Fatalf("Invalid TRUSTED_PROXIES: %v", err)
	}

	// CORS or other middleware could go here
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.RequestSizeLimitMiddleware())
	r.Use(middleware.RateLimitMiddleware())
	r.Use(middleware.IpBanMiddleware())
	r.NoRoute(func(c *gin.Context) { c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"}) })

	// Widget routes (public, no auth)
	r.GET("/bots/widget/:id/", handlers.GetBotWidget)
	r.GET("/servers/widget/:id/", handlers.GetServerWidget)

	// Badge SVG routes (public, no auth)
	r.GET("/bots/widget/:id/status.svg", handlers.GetBotStatusBadge)
	r.GET("/bots/widget/:id/servers.svg", handlers.GetBotServersBadge)
	r.GET("/bots/widget/:id/votes.svg", handlers.GetBotVotesBadge)
	r.GET("/servers/widget/:id/members.svg", handlers.GetServerMembersBadge)
	r.GET("/servers/widget/:id/votes.svg", handlers.GetServerVotesBadge)

	// Bot routes
	bots := r.Group("/bots")
	{
		// Auth required endpoints
		authorized := bots.Group("")
		authorized.Use(middleware.AuthMiddleware(), middleware.ApiUsageMiddleware())
		{
			authorized.GET("", handlers.GetBotList)
			authorized.GET("/:id", handlers.GetBotInfo)
			authorized.GET("/:id/voted", handlers.CheckBotVote)
			authorized.POST("/:id/vote", handlers.VoteBot)
			authorized.POST("/:id/stats", handlers.UpdateBotStats)
			authorized.GET("/:id/reviews", handlers.GetReviewsHandler("bot"))
			authorized.POST("/:id/reviews", handlers.CreateReviewHandler("bot"))
			authorized.PUT("/:id/reviews/:reviewId", handlers.UpdateReviewHandler("bot"))
			authorized.DELETE("/:id/reviews/:reviewId", handlers.DeleteReviewHandler("bot"))
			authorized.POST("/webhook", handlers.UpdateWebhook)
			authorized.POST("/webhook/verify", handlers.VerifyWebhook)
		}
	}

	// Server routes
	servers := r.Group("/servers")
	{
		authorized := servers.Group("")
		authorized.Use(middleware.AuthMiddleware(), middleware.ApiUsageMiddleware())
		{
			authorized.GET("", handlers.GetServerList)
			authorized.GET("/:id", handlers.GetServerInfo)
			authorized.GET("/:id/voted", handlers.CheckServerVote)
			authorized.POST("/:id/vote", handlers.VoteServer)
			authorized.GET("/:id/reviews", handlers.GetReviewsHandler("server"))
			authorized.POST("/:id/reviews", handlers.CreateReviewHandler("server"))
			authorized.PUT("/:id/reviews/:reviewId", handlers.UpdateReviewHandler("server"))
			authorized.DELETE("/:id/reviews/:reviewId", handlers.DeleteReviewHandler("server"))
			authorized.POST("/webhook", handlers.UpdateWebhook)
			authorized.POST("/webhook/verify", handlers.VerifyWebhook)
		}
	}

	// User routes
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware(), middleware.ApiUsageMiddleware())
	{
		users.GET("", handlers.GetUserList)
		users.GET("/:id", handlers.GetUserInfo)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3015"
	}
	bindAddress := os.Getenv("BIND_ADDRESS")
	if bindAddress == "" {
		bindAddress = "127.0.0.1"
	}
	server := &http.Server{
		Addr: bindAddress + ":" + port, Handler: r,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Server starting on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
