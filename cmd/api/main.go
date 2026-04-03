package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"mini-url/internal/repository"
	"mini-url/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	region := os.Getenv("AWS_REGION")
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	base62Charset := os.Getenv("BASE62_CHARSET")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dbClient := dynamodb.NewFromConfig(cfg)

	r := gin.Default()

	r.POST("/shorten", func(c *gin.Context) {
		var req struct {
			LongURL string `json:"long_url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		shortID := utils.GenerateShortID(base62Charset)
		tableName := os.Getenv("DYNAMODB_TABLE_NAME")

		err := repository.SaveURL(dbClient, tableName, shortID, req.LongURL)

		if err != nil {
			log.Printf("Error saving to DB: %v", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{
			"short_url": "http://localhost:" + port + "/" + shortID,
		})
	})

	r.GET("/:shortID", func(c *gin.Context) {
		shortID := c.Param("shortID")
		tableName := os.Getenv("DYNAMODB_TABLE_NAME")

		originalURL, err := repository.GetURL(dbClient, tableName, shortID)

		if err != nil {
			log.Printf("Error querying DB: %v", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if originalURL == "" {
			c.JSON(404, gin.H{"error": "URL not found"})
			return
		}

		c.Redirect(http.StatusFound, originalURL)
	})

	log.Printf("Server is running on port %s", port)
	r.Run(":" + port)
}
