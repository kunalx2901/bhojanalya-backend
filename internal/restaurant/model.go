package restaurant

import "time"

type Restaurant struct {
	ID          string
	Name        string
	City        string
	CuisineType string
	OwnerID     string
	Status      string
	CreatedAt   time.Time
}

type PreviewData struct {
	ID            int      `json:"id"`
	Name          string
	City          string
	CuisineType   string
	CostForTwo    float64
	Images        []string
	MenuPDFs      []string `json:"menu_pdfs"`

	Deals         []PreviewDeal `json:"deals"`
}

type PreviewDeal struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Type          string  `json:"type"`
	Category      *string `json:"category,omitempty"`
	DiscountValue float64 `json:"discount_value"`
	Status        string  `json:"status"`
}

