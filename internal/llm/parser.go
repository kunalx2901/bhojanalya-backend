package llm

import (
	"context"
	"encoding/json"
	"errors"
)

type ParsedMenu struct {
	Items []MenuItem `json:"items"`
}

type MenuItem struct {
	Name       string   `json:"name"`
	Category   *string  `json:"category"`
	Price      *float64 `json:"price"`
	Confidence float64  `json:"confidence"`
}

func ParseMenu(
	ctx context.Context,
	client Client,
	ocrText string,
) ([]MenuItem, error) {

	rawJSON, err := client.ParseMenu(ctx, ocrText)
	if err != nil {
		return nil, err
	}

	var parsed ParsedMenu
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return nil, errors.New("invalid LLM JSON output")
	}

	return parsed.Items, nil
}
