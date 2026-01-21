package llm

import (
	"context"
)

type Client interface {
	ParseMenu(ctx context.Context, ocrText string) (string, error)
}
