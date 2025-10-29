package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	geminiAPIKey  = os.Getenv("GEMINI_API_KEY")
	geminiModel   = os.Getenv("GEMINI_MODEL") //
	geminiBaseURL = "https://generativelanguage.googleapis.com/v1/models"
)

// Request format for Gemini generateContent API
type genRequest struct {
	Contents []struct {
		Role  string `json:"role"`
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
	GenerationConfig *struct {
		Temperature     float64 `json:"temperature,omitempty"`
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	} `json:"generationConfig,omitempty"`
}

// Response format for Gemini generateContent API
type genResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func GenerateCaption(ctx context.Context, imageURL string) (string, error) {
	if geminiAPIKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}
	model := geminiModel
	if model == "" {
		model = "gemini-1.5-flash"
	}
	promptText := fmt.Sprintf("Write a short friendly message describing the uploaded image located at %s. Keep it brief (1-2 sentences) and suitable for notifying subscribers about the new upload.", imageURL)

	reqBody := genRequest{
		Contents: []struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Role: "user",
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: promptText},
				},
			},
		},
		GenerationConfig: &struct {
			Temperature     float64 `json:"temperature,omitempty"`
			MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
		}{Temperature: 0.2, MaxOutputTokens: 150},
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiBaseURL, model, geminiAPIKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var rbody bytes.Buffer
		_, _ = rbody.ReadFrom(resp.Body)
		return "", fmt.Errorf("gemini API error: status %d body %s", resp.StatusCode, rbody.String())
	}

	var gr genResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", err
	}
	if len(gr.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned by gemini")
	}
	parts := gr.Candidates[0].Content.Parts
	if len(parts) == 0 {
		return "", fmt.Errorf("no parts returned by gemini")
	}
	return parts[0].Text, nil
}
