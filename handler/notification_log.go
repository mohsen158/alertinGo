package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohsen/alertinGo/db"
)

func GetNotificationLogs(c *gin.Context) {
	logs, err := db.GetNotificationLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}
