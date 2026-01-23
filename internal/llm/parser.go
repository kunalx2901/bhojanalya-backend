package llm

import (
	"encoding/json"
	"errors"
	"strings"
)

func ParseLLMResponse(raw string) (*ParsedOCRResult, error) {
	raw = strings.TrimSpace(raw)

	// Hard safety: must be pure JSON object
	if !strings.HasPrefix(raw, "{") || !strings.HasSuffix(raw, "}") {
		return nil, errors.New("LLM output is not pure JSON")
	}

	var parsed ParsedOCRResult
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	// Structural validation only
	if parsed.Items == nil {
		return nil, errors.New("missing items field in parsed JSON")
	}

	// Validate items only if present
	for _, item := range parsed.Items {
		if item.Name == "" {
			return nil, errors.New("invalid item: empty name")
		}
		if item.Price <= 0 {
			return nil, errors.New("invalid item: price must be > 0")
		}
	}

	// ✅ IMPORTANT:
	// Empty items is VALID (common for PDFs)
	return &parsed, nil
}
