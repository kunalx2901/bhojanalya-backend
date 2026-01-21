package menu

import "context"

type Repository interface {
	CreateUpload(
		restaurantID int,
		imageURL string,
		filename string,
	) (int, error)
	GetByID(ctx context.Context, id int) (*MenuUpload, error)
}
