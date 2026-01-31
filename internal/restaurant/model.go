package restaurant

import "time"

type Restaurant struct {
	ID               string
	Name             string
	City             string
	CuisineType      string
	OwnerID          string
	Status           string
	ShortDescription string
	OpensAt          string
	ClosesAt         string
	CreatedAt        time.Time
}


type PreviewData struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	City             string   `json:"city"`
	CuisineType      string   `json:"cuisine_type"`
	ShortDescription string   `json:"short_description"`
	OpensAt          string   `json:"opens_at"`
	ClosesAt         string   `json:"closes_at"`

	CostForTwo float64  `json:"cost_for_two"`
	Images     []string `json:"images"`
	MenuPDFs   []string `json:"menu_pdfs"`
	Deals      []PreviewDeal `json:"deals"`
}


type PreviewDeal struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Type          string  `json:"type"`
	Category      *string `json:"category,omitempty"`
	DiscountValue float64 `json:"discount_value"`
	Status        string  `json:"status"`
}


type AdminRestaurantDetails struct {
	ID               int
	Email            string
	OwnerName        string
	CuisineType      string
	City             string
	ShortDescription string
	OpensAt          string
	ClosesAt         string
}

