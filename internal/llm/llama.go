package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type LLaMAClient struct {
	apiKey string
	model  string
	apiURL string
}

func NewLLaMAClient() *LLaMAClient {
	return &LLaMAClient{
		apiKey: os.Getenv("LLAMA_API_KEY"),
		model:  os.Getenv("LLAMA_MODEL"),
		apiURL: os.Getenv("LLAMA_API_URL"),
	}
}

func (l *LLaMAClient) ParseMenu(ctx context.Context, ocrText string) (string, error) {
	if l.apiKey == "" {
		return "", errors.New("missing LLAMA_API_KEY")
	}

	payload := map[string]interface{}{
		"model": l.model,
		"input": buildPrompt(ocrText),
		"temperature": 0.1,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		l.apiURL,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+l.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Println("LLAMA RAW RESPONSE:", string(raw))

	// ---- Try ALL known Meta response formats ----
	rawText := string(raw)

	// TEMP DEBUG (keep for now)
	log.Println("LLAMA RAW TEXT OUTPUT:\n", rawText)

	// Extract JSON from text
	jsonText := extractJSON(rawText)
	if jsonText == "" {
		return "", errors.New("llama did not return valid JSON")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return "", err
	}

	// Variant A
	if v, ok := parsed["output_text"].(string); ok && v != "" {
		return v, nil
	}

	// Variant B
	if v, ok := parsed["generated_text"].(string); ok && v != "" {
		return v, nil
	}

	// Variant C
	if gen, ok := parsed["generation"].(map[string]interface{}); ok {
		if txt, ok := gen["text"].(string); ok && txt != "" {
			return txt, nil
		}
	}

	// If nothing matched
	return "", errors.New("empty llama response")
}

func extractJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start == -1 || end == -1 || end <= start {
		return ""
	}

	return text[start : end+1]
}
