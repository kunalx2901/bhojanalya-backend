package competition
import "time"

// Snapshot represents aggregated competitive pricing data
type Snapshot struct {
	ID                int       `json:"id"`
	City              string    `json:"city"`
	CuisineType       string    `json:"cuisine_type"`
	AvgCostForTwo     float64   `json:"avg_cost_for_two"`
	MedianCostForTwo  float64   `json:"median_cost_for_two"`
	SampleSize        int       `json:"sample_size"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}


