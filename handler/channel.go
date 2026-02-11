package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohsen/alertingGo/db"
)

type CreateChannelRequest struct {
	Name           string `json:"name" binding:"required"`
	TelegramChatID string `json:"telegram_chat_id" binding:"required"`
}

func GetChannels(c *gin.Context) {
	channels, err := db.GetAllChannels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channels)
}

func CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := db.CreateChannel(c.Request.Context(), req.Name, req.TelegramChatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

func DeleteChannel(c *gin.Context) {
	id := c.Param("id")

	if err := db.DeleteChannel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
