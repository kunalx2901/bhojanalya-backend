package menu

import (
	"errors"

	"bhojanalya/internal/llm"
)

// BuildCostForTwo builds a deterministic cost-for-two basket
// using NON-nullable Item (NOT MenuItem)
func BuildCostForTwo(parsed *llm.ParsedOCRResult) (*CostForTwo, error) {
	if parsed == nil || len(parsed.Items) == 0 {
		return nil, errors.New("empty parsed menu")
	}

	cost := &CostForTwo{
		Definition:    CostForTwoDefinition,
		SelectedItems: make(map[string][]Item),
		Confidence:    0.93,
	}

	var subtotal float64

	counts := map[string]int{
		"starter":     0,
		"main_course": 0,
		"drink":       0,
		"dessert":     0,
	}

	for _, parsedItem := range parsed.Items {
		limit, ok := CostForTwoDefinition[parsedItem.Category]
		if !ok {
			continue
		}
		if counts[parsedItem.Category] >= limit {
			continue
		}

		// ✅ USE Item (NOT MenuItem)
		item := Item{
			Name:     parsedItem.Name,
			Category: parsedItem.Category,
			Price:    parsedItem.Price,
		}

		cost.SelectedItems[parsedItem.Category] =
			append(cost.SelectedItems[parsedItem.Category], item)

		counts[parsedItem.Category]++
		subtotal += item.Price
	}

	if subtotal == 0 {
		return nil, errors.New("cost-for-two basket incomplete")
	}

	tax := subtotal * parsed.TaxPercent / 100

	cost.Calculation = CostCalculation{
		Subtotal: subtotal,
		Tax:      tax,
		Total:    subtotal + tax,
	}

	return cost, nil
}
