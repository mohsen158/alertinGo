package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohsen/alertinGo/db"
)

func CreateApiKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Generate random 32-byte key
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key"})
		return
	}
	plaintext := hex.EncodeToString(rawKey)

	hash := sha256.Sum256([]byte(plaintext))
	keyHash := hex.EncodeToString(hash[:])
	keyPrefix := plaintext[:8]

	apiKey, err := db.CreateApiKey(c.Request.Context(), req.Name, keyHash, keyPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         apiKey.ID,
		"name":       apiKey.Name,
		"key":        plaintext,
		"key_prefix": apiKey.KeyPrefix,
		"is_active":  apiKey.IsActive,
		"created_at": apiKey.CreatedAt,
		"message":    "Store this key securely. It will not be shown again.",
	})
}

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
