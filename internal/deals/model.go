package deals

type DealSuggestion struct {
	RestaurantID           int     `json:"restaurant_id"`
	City                   string  `json:"city"`
	CuisineType            string  `json:"cuisine_type"`
	Positioning            string  `json:"positioning"`
	RestaurantCostForTwo   float64 `json:"restaurant_cost_for_two"`
	MarketAvg              float64 `json:"market_avg_cost_for_two"`
	MarketMedian           float64 `json:"market_median_cost_for_two"`
	SuggestedAction        string  `json:"suggested_action"`
	SuggestedDiscount      int     `json:"suggested_discount_percent,omitempty"`
	Reason                 string  `json:"reason"`
}

type Deal struct {
	ID             int      `json:"id"`
	RestaurantID   int      `json:"restaurant_id"`
	Type           string   `json:"type"` // PERCENTAGE | FLAT | COMBO
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`

	// Optional targeting
	Category        *string `json:"category,omitempty"` // starter | main_course | drink | dessert

	// Pricing
	DiscountValue   float64 `json:"discount_value"`
	OriginalPrice   *float64 `json:"original_price,omitempty"`
	FinalPrice      *float64 `json:"final_price,omitempty"`

	// Workflow
	Status          string   `json:"status"` // DRAFT | PENDING_APPROVAL | APPROVED | REJECTED
	Suggested       bool     `json:"suggested"`

	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

