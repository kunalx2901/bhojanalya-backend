package restaurant

type CompetitiveInsight struct {
	RestaurantID         int     `json:"restaurant_id"`
	City                 string  `json:"city"`
	CuisineType          string  `json:"cuisine_type"`
	RestaurantCostForTwo float64 `json:"restaurant_cost_for_two"`
	MarketAvg            float64 `json:"market_avg"`
	MarketMedian         float64 `json:"market_median"`
	SampleSize           int     `json:"sample_size"`
	Positioning          string  `json:"positioning"`
}
