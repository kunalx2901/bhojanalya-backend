package menu

// ParsedMenu is the validated, normalized menu
// used by pricing, deals, and competitive insights
type ParsedMenu struct {
	Items      []Item  `json:"items"`
	TaxPercent float64 `json:"tax_percent"`
}

// MenuUpload represents a parsed menu waiting for admin approval
// Used only by ADMIN flows
type MenuUpload struct {
	ID           int                    `json:"id"`
	RestaurantID int                    `json:"restaurant_id"`

	// Restaurant context
	RestaurantName string `json:"restaurant_name"`
	City           string `json:"city"`
	CuisineType    string `json:"cuisine_type"`
	OpensAt        string `json:"opens_at"`
	ClosesAt       string `json:"closes_at"`

	// Menu data
	Filename   string                 `json:"filename"`
	ParsedData map[string]interface{} `json:"parsed_data"`
}
