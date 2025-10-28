package config

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sachinggsingh/notify/internal/models"
)

var UpgraderWs = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins
	},
}

var HubInstance = &models.Hub{
	Clients:    make(map[*models.Client]bool),
	Broadcast:  make(chan []byte),
	Register:   make(chan *models.Client),
	Unregister: make(chan *models.Client),
	Mutex:      sync.RWMutex{},
}

func init() {
	go HubInstance.Run()
}

func HandleConnection() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, exits := c.Get("uid")
		if !exits {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
			return
		}

		userIdStr := userId.(string)
		conn, err := UpgraderWs.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		client := &models.Client{
			ID:       userIdStr,
			UserID:   userIdStr,
			Conn:     conn,
			Send:     make(chan []byte, 256),
			Hub:      HubInstance,
			LastPing: time.Now(),
		}
		client.Hub.Register <- client

		// Start goroutines for reading and writing
		go client.WritePump()
		go client.ReadPump()
		go client.StartPingPong()
	}
}

func HandleDisconnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := c.MustGet("client").(*models.Client)
		client.Hub.Unregister <- client
		client.Conn.Close()
	}
}
