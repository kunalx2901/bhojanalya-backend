package llm

type ParsedItem struct {
	Name     string  `json:"name"`
	Category string  `json:"category"` // starter | main_course | drink | dessert
	Price    float64 `json:"price"`
}

type ParsedOCRResult struct {
	Items      []ParsedItem `json:"items"`
	TaxPercent float64      `json:"tax_percent"`
}
