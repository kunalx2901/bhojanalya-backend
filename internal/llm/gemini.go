package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type GeminiClient struct {
	apiKey string
	model  string
}

func NewGeminiClient() *GeminiClient {
	return &GeminiClient{
		apiKey: os.Getenv("GEMINI_API_KEY"),
		model:  os.Getenv("GEMINI_MODEL"),
	}
}

// ParseOCR sends OCR raw text to Gemini and guarantees JSON-only output
func (g *GeminiClient) ParseOCR(ctx context.Context, ocrText string) (string, error) {
	if g.apiKey == "" {
		return "", errors.New("missing GEMINI_API_KEY")
	}
	if g.model == "" {
		return "", errors.New("missing GEMINI_MODEL")
	}
	if ocrText == "" {
		return "", errors.New("empty OCR text")
	}

	prompt := BuildOCRParsePrompt(ocrText)

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model,
		g.apiKey,
	)

	payload := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature":     0.2,
			"maxOutputTokens": 2048,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 🔥 Keep this during development
	fmt.Println("GEMINI RAW RESPONSE:", string(raw))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini api error: %s", string(raw))
	}

	// Gemini response shape
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 ||
		len(result.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty gemini response")
	}

	output := result.Candidates[0].Content.Parts[0].Text

	// 🔒 CRITICAL: Ensure JSON-only output
	if !json.Valid([]byte(output)) {
		return "", errors.New("gemini returned non-json output")
	}

	return output, nil
}
