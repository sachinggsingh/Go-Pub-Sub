package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Err          error  `json:"err"`
}

type MessageResponse struct {
	Message string `json:"message"`
	Err     error  `json:"err"`
}

type ImageResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	PublicID string `json:"public_id"`
	ImageID  string `json:"image_id"`
	Err      error  `json:"err"`
}

type ErrorResponse struct {
	Err error `json:"err"`
}

func main() {
	baseURl := "http://localhost:8080"

	fmt.Println("=== Step 1: User Login ===")

	token := loginUser(baseURl)
	if token == "" {
		fmt.Println("Failed to login user")
		return
	}

	fmt.Println("Login Successfully ", token[:20]+"...")

	fmt.Println("=== Step 2: Upload an image (protected) ===")
	img, upErr := uploadImage(baseURl, token)
	if upErr != nil {
		log.Printf("Upload failed: %v\n", upErr)
	} else {
		log.Printf("Upload succeeded: id=%s url=%s\n", img.ImageID, img.URL)
	}

	fmt.Println("=== Step 3: Testing REST API Endpoints ===")
	// Test publish (protected)
	if err := publishMessage(baseURl, token, "demo-user", "Hello from test client"); err != nil {
		log.Printf("Publish failed: %v\n", err)
	} else {
		log.Println("Publish succeeded")
	}

	fmt.Println("=== Step 4: Testing WebSocket pub/sub ===")
	// Connect to WS and, once connected, publish a message referencing the uploaded image
	onConnect := func() {
		content := "WS publish after upload"
		if img.ImageID != "" {
			content = fmt.Sprintf("WS publish after upload image_id=%s", img.ImageID)
		}
		if err := publishMessage(baseURl, token, "demo-user", content); err != nil {
			log.Printf("Publish (during WS) failed: %v\n", err)
		} else {
			log.Println("Publish (during WS) succeeded")
		}
	}

	// Connect to WS and listen briefly
	if err := testWebSocket("ws://localhost:8080/protected/ws", token, 10*time.Second, onConnect); err != nil {
		log.Printf("WebSocket test failed: %v\n", err)
	} else {
		log.Println("WebSocket test completed")
	}
}

func loginUser(baseUrl string) string {
	loginData := map[string]string{
		"email":    "sachingajendrasingh@gmail.com",
		"password": "123456789",
	}

	jsonData, marshalErr := json.Marshal(loginData)
	if marshalErr != nil {
		fmt.Println("Error in marshalling")
	}

	resp, postErr := http.Post(baseUrl+"/user/login", "application/json", bytes.NewBuffer(jsonData))
	if postErr != nil {
		fmt.Printf("Login request failed: %v\n", postErr)
		return ""
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Login failed with status %d: %s\n", resp.StatusCode, string(body))
		return ""
	}
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		fmt.Printf("Failed to decode login response: %v\n", err)
		return ""
	}

	return loginResp.Token
}

// publishMessage calls POST /protected/publish with Authorization header
func publishMessage(baseUrl, token, userID, content string) error {
	payload := map[string]any{
		"user_id":   userID,
		"content":   content,
		"timestamp": time.Now(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseUrl+"/protected/publish", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}
	log.Printf("Publish response: %s\n", string(b))
	return nil
}

// testWebSocket dials the WS endpoint with Authorization header and listens for messages for duration
func testWebSocket(rawURL, token string, listenFor time.Duration, onConnected func()) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer c.Close()

	log.Println("WebSocket connected")

	if onConnected != nil {
		go onConnected()
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("read error: %v\n", err)
				return
			}
			log.Printf("recv: %s\n", message)
		}
	}()

	// Optionally send a hello message (server currently ignores content)
	if err := c.WriteMessage(websocket.TextMessage, []byte("hello from client")); err != nil {
		log.Printf("write error: %v\n", err)
	}

	select {
	case <-time.After(listenFor):
		log.Println("WS listen window elapsed; closing")
	case <-done:
		log.Println("WS reader exited early")
	}
	return nil
}

// uploadImage sends a multipart/form-data POST to /protected/upload with an in-memory tiny PNG
func uploadImage(baseUrl, token string) (ImageResponse, error) {
	var result ImageResponse

	// Create the form file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create a sample image file part
	fw, err := writer.CreateFormFile("image", "test-image.jpg")
	if err != nil {
		return result, fmt.Errorf("create form file: %w", err)
	}

	// Write some sample image data (you can replace this with a real image file)
	_, err = fw.Write([]byte("sample image content"))
	if err != nil {
		return result, fmt.Errorf("write file content: %w", err)
	}

	writer.Close()

	// Create the request
	req, err := http.NewRequest(http.MethodPost, baseUrl+"/protected/upload", &buf)
	if err != nil {
		return result, fmt.Errorf("new request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response directly into our ImageResponse struct
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
