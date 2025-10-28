package models

import (
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email         *string            `json:"email" validate:"required,email"`
	Password      *string            `json:"password" validate:"required,min=8,max=24"`
	Token         *string            `json:"token"`
	Refresh_Token *string            `json:"refresh_token"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
	User_id       string             `json:"user_id" bson:"user_id"`
}

// UploadData model
type Image struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	URL      string             `bson:"url" json:"url"`
	PublicID string             `bson:"public_id" json:"public_id"`
	Uploaded time.Time          `bson:"uploaded" json:"uploaded"`
	Image_id string             `json:"image_id" validate:"required"`
}

// TokenResponse model
type TokenResponse struct {
	Token        string
	RefreshToken string
	Err          error
}

// JWT claims
type SignedDetails struct {
	Email  string
	UserID string
	jwt.RegisteredClaims
}

type Message struct {
	UserID    string         `json:"user_id" validate:"required"`
	Content   string         `json:"content" validate:"required"`
	Timestamp time.Time      `json:"timestamp"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type Client struct {
	ID       string
	UserID   string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	LastPing time.Time
}

type Hub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	Mutex      sync.RWMutex
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client] = true
			h.Mutex.Unlock()
			println("Client connected:", client.ID)

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.Mutex.Unlock()
			println("Client disconnected:", client.ID)

		case message := <-h.Broadcast:
			h.Mutex.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.Mutex.RUnlock()
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.LastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (c *Client) StartPingPong() {
	for {
		time.Sleep(10 * time.Second)
		if time.Since(c.LastPing) > 30*time.Second {
			c.Hub.Unregister <- c
			c.Conn.Close()
			break
		}
	}
}
