package llm

import (
	"encoding/json"
	"errors"
)

func ParseLLMResponse(raw string) (*ParsedOCRResult, error) {
	var parsed ParsedOCRResult

	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	if len(parsed.Items) == 0 {
		return nil, errors.New("no items parsed from OCR")
	}

	for _, item := range parsed.Items {
		if item.Name == "" || item.Price <= 0 {
			return nil, errors.New("invalid item detected")
		}
	}

	return &parsed, nil
}
