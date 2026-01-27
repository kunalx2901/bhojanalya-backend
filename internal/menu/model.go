package menu

import "time"

// ParsedMenu is the validated, normalized menu
// used by pricing, deals, and competitive insights
type ParsedMenu struct {
	Items      []Item  `json:"items"`
	TaxPercent float64 `json:"tax_percent"`
}

// MenuAnalysis matches the JSON structure you want from Llama-3.1-8B
type MenuAnalysis struct {
	Categories []CategoryDetail `json:"categories"`
}

type CategoryDetail struct {
	Category   string  `json:"category"`
	Cuisine    string  `json:"cuisine,omitempty"` // Added for your cuisine requirement
	YourAvg    float64 `json:"your_avg"`
	MarketAvg  float64 `json:"market_avg"`
	Difference float64 `json:"difference"`
	Status     string  `json:"status"` // e.g., "Overpriced" or "Fair"
}

// MenuUpload represents the database row in menu_uploads
type MenuUpload struct {
	ID             int          `json:"id"`
	RestaurantID   int          `json:"restaurant_id"`
	ImageURL       string       `json:"image_url"`
	RawText        string       `json:"raw_text,omitempty"`
	StructuredData MenuAnalysis `json:"structured_data,omitempty"`
	Status         UploadStatus `json:"status"`
	OCRError       *string      `json:"ocr_error,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}