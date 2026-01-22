package menu

type Repository interface {
	CreateUpload(
		restaurantID int,
		imageURL string,
		filename string,
	) (int, error)

	SaveMenuItems(
		menuUploadID int,
		restaurantID int,
		items []MenuItem,
	) error

	SaveParsedMenu(
		menuUploadID int,
		doc map[string]interface{},
	) error
}
