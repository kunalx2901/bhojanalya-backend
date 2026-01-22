package menu

// CostForTwoDefinition defines the fixed comparison basket
var CostForTwoDefinition = map[string]int{
	"starter":     1,
	"main_course": 1,
	"drink":       2,
	"dessert":     1,
}

// CostForTwo represents a standardized pricing basket
type CostForTwo struct {
	Definition    map[string]int      `json:"definition"`
	SelectedItems map[string][]Item   `json:"selected_items"` // 🔥 Item, NOT MenuItem
	Calculation   CostCalculation     `json:"calculation"`
	Confidence    float64             `json:"confidence"`
}

type CostCalculation struct {
	Subtotal float64 `json:"subtotal"`
	Tax      float64 `json:"tax"`
	Total    float64 `json:"total_cost_for_two"`
}
