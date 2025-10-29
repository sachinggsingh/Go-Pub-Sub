package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	config "github.com/sachinggsingh/notify/internal/config"
	database "github.com/sachinggsingh/notify/internal/db"
	"github.com/sachinggsingh/notify/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var messageCollection *mongo.Collection = database.OpenCollection(database.Client, "message")
var validate = validator.New()

func PublishMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var message models.Message

		if err := c.BindJSON(&message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(message)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, err := messageCollection.InsertOne(ctx, message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
			return
		}

		data, err := json.Marshal(message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal message"})
			return
		}

		err = config.RDB.Publish(ctx, "message", data).Err()
		if err != nil {
			log.Println("Redis publish error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish message"})
			return
		}

		c.JSON(http.StatusOK, bson.M{
			"message": "Message published successfully",
			"result":  result,
		})
	}
}

// StartRedisSubscriber runs a background subscriber to persist messages
func StartRedisSubscriber(ctx context.Context) {
	sub := config.RDB.Subscribe(ctx, "message")
	ch := sub.Channel()
	for msg := range ch {
		var message models.Message
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}
		// Save to MongoDB
		saveCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := messageCollection.InsertOne(saveCtx, message)
		cancel()
		if err != nil {
			log.Println("Failed to insert message:", err)
			continue
		}
		log.Println("Message received and saved:", message)

		// Broadcast to all connected WebSocket clients
		if b, err := json.Marshal(message); err == nil {
			config.HubInstance.Broadcast <- b
		} else {
			log.Println("Failed to marshal message for WS broadcast:", err)
		}
	}
}
