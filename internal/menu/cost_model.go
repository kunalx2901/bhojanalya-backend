package menu

// CostForTwo represents a standardized, comparable pricing signal
type CostForTwo struct {
	Availability map[string]bool `json:"availability"`
	Selected     map[string]float64 `json:"selected"`
	Calculation  CostCalculation `json:"calculation"`
	Confidence   float64         `json:"confidence"`
}

type CostCalculation struct {
	Subtotal float64 `json:"subtotal"`
	Tax      float64 `json:"tax"`
	Total    float64 `json:"total_cost_for_two"`
}
