package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohsen/alertinGo/db"
)

func GetMonitors(c *gin.Context) {
	monitors, err := db.GetAllMonitors(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, monitors)
}

func GetMonitor(c *gin.Context) {
	id := c.Param("id")

	monitor, err := db.GetMonitorByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
		return
	}

	c.JSON(http.StatusOK, monitor)
}

type UpdateMonitorRequest struct {
	IsActive  *bool   `json:"is_active"`
	ChannelID *string `json:"channel_id"`
}

func UpdateMonitor(c *gin.Context) {
	id := c.Param("id")

	var req UpdateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current monitor first
	existing, err := db.GetMonitorByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
		return
	}

	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	channelID := existing.ChannelID
	if req.ChannelID != nil {
		channelID = req.ChannelID
	}

	monitor, err := db.UpdateMonitor(c.Request.Context(), id, isActive, channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, monitor)
}

func DeleteMonitor(c *gin.Context) {
	id := c.Param("id")

	if err := db.DeleteMonitor(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
