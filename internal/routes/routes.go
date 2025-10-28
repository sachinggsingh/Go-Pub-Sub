package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/sachinggsingh/notify/internal/config"
	controller "github.com/sachinggsingh/notify/internal/controllers"
	"github.com/sachinggsingh/notify/internal/middleware"
)

func UserRoutes(incommingRoutes *gin.Engine) {
	incommingRoutes.POST("/user/signup", controller.CreateUser())
	incommingRoutes.POST("/user/login", controller.Login())
}

func ImageRoutes(incommingRoutes *gin.Engine) {

	protectedRoutes := incommingRoutes.Group("/protected")
	protectedRoutes.Use(middleware.Authenticate())
	{
		protectedRoutes.GET("/image/:image_id", controller.GetImage())
		protectedRoutes.POST("/upload", controller.UploadFile())
	}
}

func WebsocketRoutes(incommingRoutes *gin.Engine) {
	protectedRoutes := incommingRoutes.Group("/protected")
	protectedRoutes.Use(middleware.Authenticate())
	{
		protectedRoutes.GET("/ws", config.HandleConnection())
		protectedRoutes.GET("/ws/disconnect", config.HandleDisconnect())
	}
}

func PubSubRoutes(incommingRoutes *gin.Engine) {
	protectedRoutes := incommingRoutes.Group("/protected")
	protectedRoutes.Use(middleware.Authenticate())
	{
		protectedRoutes.POST("/publish", controller.PublishMessage())
	}
}
