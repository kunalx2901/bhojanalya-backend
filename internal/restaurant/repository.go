package restaurant

import "context"

type Repository interface {
	// core
	Create(restaurant *Restaurant) error
	ListByOwner(ownerID string) ([]*Restaurant, error)

	// ownership & competitive insights
	IsOwner(ctx context.Context, restaurantID int, userID string) (bool, error)
	GetLatestParsedCostForTwo(
		ctx context.Context,
		restaurantID int,
	) (float64, string, string, error)
}
