package menu

// Item is the normalized, non-nullable menu item
// used ONLY for pricing, deals, and insights
type Item struct {
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Price      float64 `json:"price"`
	Confidence float64 `json:"confidence,omitempty"`
}
