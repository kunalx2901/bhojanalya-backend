package llm

func BuildOCRParsePrompt(ocrText string) string {
	return `
You are a data extraction engine.

Your task:
- Convert the OCR text into STRICT JSON.
- Output MUST be valid JSON.
- Output MUST start with { and end with }.
- Output MUST contain ONLY JSON.
- NO explanations.
- NO markdown.
- NO comments.
- NO extra text.

If you cannot extract data, return this exact JSON:
{
  "items": [],
  "tax_percent": 0
}

Required JSON schema:
{
  "items": [
    {
      "name": "string",
      "category": "starter | main_course | drink | dessert",
      "price": number
    }
  ],
  "tax_percent": number
}

OCR TEXT:
` + ocrText
}
