package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	config "github.com/sachinggsingh/notify/internal/config"
	controller "github.com/sachinggsingh/notify/internal/controllers"
	routes "github.com/sachinggsingh/notify/internal/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	gin_mode := os.Getenv("GIN_MODE")
	if gin_mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		fmt.Println("Welcome to the Restaurant Management System")
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.Use(gin.Logger())
	config.InitCloudinary()
	if err := config.InitRedis(); err != nil {
		log.Fatal(err)
	}

	// Register routes (auth applied within route groups)
	routes.UserRoutes(router)
	routes.ImageRoutes(router)
	routes.WebsocketRoutes(router)
	routes.PubSubRoutes(router)

	// Start background Redis subscriber
	go controller.StartRedisSubscriber(context.Background())

	router.Run(":" + port)

}
