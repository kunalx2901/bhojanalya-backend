package llm

func BuildOCRParsePrompt(ocrText string) string {
	return `
You are extracting structured data from restaurant OCR text.

Rules:
- Extract ONLY item names, categories, and prices.
- Categories MUST be one of: starter, main_course, drink, dessert.
- Do NOT calculate totals.
- Do NOT infer missing prices.
- If tax percentage is mentioned, extract it.
- Output ONLY valid JSON.
- NO markdown, NO explanation.

Required JSON format:
{
  "items": [
    { "name": "", "category": "", "price": 0 }
  ],
  "tax_percent": 0
}

OCR TEXT:
` + ocrText
}
