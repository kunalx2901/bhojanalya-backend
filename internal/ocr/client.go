package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type ParserClient struct {
	Token   string
	BaseURL string
	client  *http.Client
}

func NewParserClient(token string) *ParserClient {
	return &ParserClient{
		Token:   token,
		BaseURL: "https://router.huggingface.co/hf-inference/models/mistralai/Mistral-7B-Instruct-v0.2",
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *ParserClient) ParseWithLogic(rawText string) (string, error) {
	prompt := fmt.Sprintf(
		`You are a system that converts restaurant menu text into valid JSON.

Rules:
- Group items by cuisine
- Market average for mains is 260
- If price > 260 → status = "Overpriced"
- Output ONLY valid JSON, no explanation

Menu text:
%s
`, rawText)

	payload, err := json.Marshal(map[string]interface{}{
		"inputs": prompt,
		"parameters": map[string]interface{}{
			"max_new_tokens": 800,
			"temperature":   0.2,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HF request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("HF API error - Status: %d, Response: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("HF API returned status %d: %s", resp.StatusCode, string(body))
	}

	var hfResp []map[string]interface{}
	if err := json.Unmarshal(body, &hfResp); err != nil {
		return "", fmt.Errorf("failed to parse HF response: %w", err)
	}

	if len(hfResp) == 0 {
		return "", fmt.Errorf("empty HF response")
	}

	text, ok := hfResp[0]["generated_text"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected HF response format")
	}

	return text, nil
}
