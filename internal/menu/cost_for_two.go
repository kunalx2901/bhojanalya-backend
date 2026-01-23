package menu

import "errors"

// BuildCostForTwo builds a deterministic cost-for-two basket
// PURE business logic (NO llm / NO ocr)
func BuildCostForTwo(menu *ParsedMenu) (*CostForTwo, error) {
	if menu == nil || len(menu.Items) == 0 {
		return nil, errors.New("empty parsed menu")
	}

	cost := &CostForTwo{
		Availability: make(map[string]bool),
		Selected:     make(map[string]float64),
		Confidence:   0.93,
	}

	required := map[string]int{
		"starter":     1,
		"main_course": 1,
		"drink":       2,
		"dessert":     1,
	}

	counts := map[string]int{}
	var subtotal float64

	for _, item := range menu.Items {
		limit, ok := required[item.Category]
		if !ok {
			continue
		}
		if counts[item.Category] >= limit {
			continue
		}

		cost.Selected[item.Category] += item.Price
		counts[item.Category]++
		subtotal += item.Price
	}

	// availability flags
	for k, v := range required {
		cost.Availability[k] = counts[k] >= v
	}

	if subtotal == 0 {
		return nil, errors.New("insufficient data to calculate cost for two")
	}

	tax := subtotal * menu.TaxPercent / 100

	cost.Calculation.Subtotal = subtotal
	cost.Calculation.Tax = tax
	cost.Calculation.Total = subtotal + tax

	return cost, nil
}
