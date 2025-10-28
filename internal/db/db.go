package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	mongo_uri := os.Getenv("MONGODB_URI")
	if mongo_uri == "" {
		mongo_uri = "mongodb://localhost:27017/upload-notify"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongo_uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to DB")
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

var Client *mongo.Client = ConnectDB()

func OpenCollection(cient *mongo.Client, collectionName string) *mongo.Collection {
	return cient.Database("notify").Collection(collectionName)
}
