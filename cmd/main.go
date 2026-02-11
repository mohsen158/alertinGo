package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mohsen/alertinGo/db"
	"github.com/mohsen/alertinGo/handler"
	"github.com/mohsen/alertinGo/middleware"
	"github.com/mohsen/alertinGo/watcher"
)

func main() {
	_ = godotenv.Load()

	db.Connect()
	db.RunMigrations()

	watcher.Start()

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		api.POST("/heartbeat", middleware.RequireAPIKey(), handler.PostHeartbeat)

		api.GET("/api-keys", handler.ListApiKeys)
		api.POST("/api-keys", handler.CreateApiKey)
		api.DELETE("/api-keys/:id", handler.DeleteApiKey)

		api.GET("/monitors", handler.GetMonitors)
		api.GET("/monitors/:id", handler.GetMonitor)
		api.PUT("/monitors/:id", handler.UpdateMonitor)
		api.DELETE("/monitors/:id", handler.DeleteMonitor)

		api.GET("/channels", handler.GetChannels)
		api.POST("/channels", handler.CreateChannel)
		api.DELETE("/channels/:id", handler.DeleteChannel)

		api.GET("/notification-logs", handler.GetNotificationLogs)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
