package core

import "context"

type RestaurantReader interface {
	IsOwner(ctx context.Context, restaurantID int, userID string) (bool, error)

	GetLatestParsedCostForTwo(
		ctx context.Context,
		restaurantID int,
	) (float64, string, string, error)
}
