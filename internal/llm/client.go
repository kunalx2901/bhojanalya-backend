package llm

import "context"

type Client interface {
	ParseOCR(ctx context.Context, ocrText string) (*ParsedOCRResult, error)
}
