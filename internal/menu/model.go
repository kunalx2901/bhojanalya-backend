package menu

type UploadStatus string

const (
	StatusMenuUploaded UploadStatus = "MENU_UPLOADED"
	StatusFailed       UploadStatus = "FAILED"
)

type MenuItem struct {
	Name       string  `json:"name"`
	Category   *string `json:"category"`
	Price      *float64 `json:"price"`
	Confidence float64 `json:"confidence,omitempty"`
}
