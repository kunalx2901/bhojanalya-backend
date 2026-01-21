package llm

import (
	"github.com/gin-gonic/gin"
)

func TestLLaMA(llama Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		testOCR := `
Pizza Margherita
Pasta Alfredo
$10
$12
`

		result, err := llama.ParseMenu(c.Request.Context(), testOCR)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"llm_output": result,
		})
	}
}
