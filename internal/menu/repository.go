package menu

type Repository interface {
	CreateUpload(
		restaurantID int,
		imageURL string,
		filename string,
	) (int, error)
}
