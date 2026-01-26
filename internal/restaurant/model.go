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


