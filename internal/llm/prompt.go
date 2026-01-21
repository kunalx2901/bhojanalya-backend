package llm

// func buildPrompt(ocrText string) string {
// 	return `
// You are a strict data parser.

// Rules:
// - Return ONLY valid JSON.
// - Do not include explanations.
// - Do not include markdown.
// - If output is not JSON, it is INVALID.
// - Convert OCR menu text into structured JSON
// - Extract item names and prices
// - Match items to prices correctly
// - Infer category if possible
// - If price missing, set price = null
// - DO NOT guess prices
// - DO NOT add explanations
// - DO NOT add markdown
// - Return ONLY valid JSON

// Output format:
// {
//   "items": [
//     {
//       "name": string,
//       "category": string | null,
//       "price": number | null,
//       "confidence": number
//     }
//   ]
// }

// OCR TEXT:
// ` + ocrText
// }


func buildPrompt(_ string) string {
	return "Say only the word HELLO and nothing else."
}
