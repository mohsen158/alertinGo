package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohsen/alertinGo/db"
)

func ListApiKeys(c *gin.Context) {
	keys, err := db.GetAllApiKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list API keys"})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func DeleteApiKey(c *gin.Context) {
	id := c.Param("id")
	if err := db.DeleteApiKey(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete API key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "API key deleted"})
}
