package menu

import (
	"errors"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// to upload a new menu for a restaurant
func (s *Service) UploadMenu(restaurantID string, filePath string) error {
	if restaurantID == "" || filePath == "" {
		return errors.New("invalid menu upload")
	}

	// Validate file type
	allowed := []string{".pdf", ".txt", ".csv", ".json", ".xml"}
	valid := false
	for _, ext := range allowed {
		if strings.HasSuffix(filePath, ext) {
			valid = true
			break
		}
	}

	if !valid {
		return errors.New("unsupported storage format")
	}

	menu := &Menu{
		RestaurantID: restaurantID,
		FilePath:     filePath,
	}

	return s.repo.Create(menu)
}

