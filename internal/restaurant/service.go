package restaurant

import "errors"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}


// core logic for creating a restaurant
func (s *Service) CreateRestaurant(
	name string,
	city string,
	cuisineType string,
	ownerID string,
) (*Restaurant, error) {

	// 🔒 Validation
	if name == "" || city == "" || cuisineType == "" {
		return nil, errors.New("missing required fields")
	}

	restaurant := &Restaurant{
		Name:        name,
		City:        city,
		CuisineType: cuisineType,
		OwnerID:     ownerID,
		Status:      "pending", // 🚨 Always pending initially
	}

	if err := s.repo.Create(restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

// lists restaurants for a specific owner
func (s *Service) ListMyRestaurants(ownerID string) ([]*Restaurant, error) {
	return s.repo.ListByOwner(ownerID)
}

