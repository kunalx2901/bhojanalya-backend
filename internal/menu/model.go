package menu

import "time"

type Menu struct {
	ID           string
	RestaurantID string
	FilePath     string
	UploadedAt   time.Time
}
