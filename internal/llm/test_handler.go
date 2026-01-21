package llm

import (
	"github.com/gin-gonic/gin"
)

func TestGeminiHandler(c *gin.Context) {
	client := NewGeminiClient()

	output, err := client.ParseText(
		c.Request.Context(),
		"Say HELLO and nothing else.",
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"output": output,
	})
}
