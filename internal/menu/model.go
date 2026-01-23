package menu

// ParsedMenu is the validated, normalized menu
// used by pricing, deals, and competitive insights
type ParsedMenu struct {
	Items      []Item  `json:"items"`
	TaxPercent float64 `json:"tax_percent"`
}
