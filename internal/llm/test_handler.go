package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func TestOCRParsing() (*ParsedOCRResult, error) {
	client := NewGeminiClient()

	ocrText := `
Paneer Tikka 220
Butter Chicken 320
Cold Drink 90
Cold Drink 90
Gulab Jamun 120
GST 10%
`

	// ✅ Call ParseOCR directly (prompt handled internally)
	rawJSON, err := client.ParseOCR(context.Background(), ocrText)
	if err != nil {
		return nil, err
	}

	parsed, err := ParseLLMResponse(rawJSON)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Parsed OCR Result:\n%+v\n", parsed)
	return parsed, nil
}

// DEV-ONLY test endpoint
func TestGeminiHandler(c *gin.Context) {
	result, err := TestOCRParsing()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
