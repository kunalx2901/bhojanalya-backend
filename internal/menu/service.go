package menu

import "errors"

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

	menu := &Menu{
		RestaurantID: restaurantID,
		FilePath:     filePath,
	}

	return s.repo.Create(menu)
}

