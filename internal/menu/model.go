package menu

type UploadStatus string

const (
	StatusMenuUploaded UploadStatus = "MENU_UPLOADED"
	StatusFailed       UploadStatus = "FAILED"
)

type MenuItem struct {
	Name       string
	Category   *string
	Price      *float64
	Confidence float64
}
