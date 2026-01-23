package menu

type Repository interface {
	// Create a menu upload entry (raw menu file)
	CreateUpload(
		restaurantID int,
		objectKey string,   // R2 object key (NOT public URL)
		filename string,
	) (int, error)

	// Save parsed menu + cost-for-two as JSON
	SaveParsedMenu(
		menuUploadID int,
		doc map[string]interface{},
	) error
}
