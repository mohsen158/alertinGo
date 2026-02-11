package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohsen/alertingGo/db"
	"github.com/mohsen/alertingGo/model"
)

type HeartbeatRequest struct {
	MonitorName     string      `json:"monitor_name" binding:"required"`
	CheckType       string      `json:"check_type" binding:"required"`
	Message         string      `json:"message"`
	Metadata        interface{} `json:"metadata"`
	Timeout         int         `json:"timeout"`
	ReAlertInterval int         `json:"re_alert_interval"`
	ServerIP        string      `json:"server_ip"`
	ServerName      string      `json:"server_name"`
}

func PostHeartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Timeout <= 0 {
		req.Timeout = 60
	}
	if req.ReAlertInterval <= 0 {
		req.ReAlertInterval = 300
	}

	metadataStr := "{}"
	if req.Metadata != nil {
		b, err := json.Marshal(req.Metadata)
		if err == nil {
			metadataStr = string(b)
		}
	}

	m := &model.Monitor{
		MonitorName:     req.MonitorName,
		CheckType:       req.CheckType,
		Message:         req.Message,
		Metadata:        metadataStr,
		Timeout:         req.Timeout,
		ReAlertInterval: req.ReAlertInterval,
		ServerIP:        req.ServerIP,
		ServerName:      req.ServerName,
	}

	result, err := db.UpsertMonitor(c.Request.Context(), m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
