package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/mohsen/alertinGo/db"
)

func main() {
	name := flag.String("name", "", "API key name (required)")
	flag.Parse()

	if *name == "" {
		log.Fatal("--name is required")
	}

	_ = godotenv.Load()

	db.Connect()
	db.RunMigrations()

	ctx := context.Background()

	// Generate a random 32-byte API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		log.Fatalf("failed to generate API key: %v", err)
	}
	plainKey := hex.EncodeToString(keyBytes)

	// Hash it with SHA-256
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])
	keyPrefix := plainKey[:8]

	// Store the hashed key
	apiKey, err := db.CreateApiKey(ctx, *name, keyHash, keyPrefix)
	if err != nil {
		log.Fatalf("failed to create API key: %v", err)
	}

	fmt.Println("API key created successfully!")
	fmt.Printf("  ID:      %s\n", apiKey.ID)
	fmt.Printf("  Name:    %s\n", apiKey.Name)
	fmt.Printf("  API Key: %s\n", plainKey)
	fmt.Println()
	fmt.Println("Save this API key — it will not be shown again.")
}
