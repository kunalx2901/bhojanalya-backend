package deals

import "time"

// --------------------------------------------------
// DEAL SUGGESTION (READ-ONLY)
// --------------------------------------------------

type DealSuggestion struct {
	RestaurantID         int     `json:"restaurant_id"`
	City                 string  `json:"city"`
	CuisineType          string  `json:"cuisine_type"`
	Positioning          string  `json:"positioning"`
	RestaurantCostForTwo float64 `json:"restaurant_cost_for_two"`
	MarketAvg            float64 `json:"market_avg_cost_for_two"`
	MarketMedian         float64 `json:"market_median_cost_for_two"`

	Suggestions []SuggestedDeal `json:"suggestions"`
}

type SuggestedDeal struct {
	Type          string   `json:"type"`     // PERCENTAGE | FLAT | COMBO
	Category      *string  `json:"category"` // starter | main_course | drink | dessert
	Title         string   `json:"title"`
	DiscountValue float64  `json:"discount_value"`
	Reason        string   `json:"reason"`
}

// --------------------------------------------------
// DEAL (PERSISTED ENTITY)
// --------------------------------------------------

type Deal struct {
	ID           int        `json:"id"`
	RestaurantID int        `json:"restaurant_id"`

	Type        string     `json:"type"`   // PERCENTAGE | FLAT | COMBO
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`

	Category     *string    `json:"category,omitempty"`

	DiscountValue float64   `json:"discount_value"`
	OriginalPrice *float64  `json:"original_price,omitempty"`
	FinalPrice    *float64  `json:"final_price,omitempty"`

	Status     string     `json:"status"` // DRAFT | PENDING_APPROVAL | APPROVED | REJECTED
	Suggested  bool       `json:"suggested"`

	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
