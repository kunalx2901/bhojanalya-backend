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

	// preview support
	HasAnyDeal(ctx context.Context, restaurantID int) (bool, error)
	GetPreviewData(ctx context.Context, restaurantID int) (*PreviewData, error)

	// restaurant images
	SaveRestaurantImages(ctx context.Context, restaurantID int, images []string) error
	GetRestaurantImages(ctx context.Context, restaurantID int) ([]string, error)

	// admin views
	ListApproved(ctx context.Context) ([]*Restaurant, error)
	GetAdminDetails(
		ctx context.Context,
		restaurantID int,
	) (*AdminRestaurantDetails, error)

}
