package restaurant

import (
	"context"
	"errors"
	"bhojanalya/internal/competition"
)

type Service struct {
	repo            Repository
	competitionRepo *competition.Repository
}

func NewService(
	repo Repository,
	competitionRepo *competition.Repository,
) *Service {
	return &Service{
		repo:            repo,
		competitionRepo: competitionRepo,
	}
}

// --------------------------------------------------
// Create restaurant
// --------------------------------------------------
func (s *Service) CreateRestaurant(
	name string,
	city string,
	cuisineType string,
	ownerID string,
) (*Restaurant, error) {

	if name == "" || city == "" || cuisineType == "" {
		return nil, errors.New("missing required fields")
	}

	restaurant := &Restaurant{
		Name:        name,
		City:        city,
		CuisineType: cuisineType,
		OwnerID:     ownerID,
		Status:      "pending",
	}

	if err := s.repo.Create(restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

// --------------------------------------------------
// List restaurants owned by user
// --------------------------------------------------
func (s *Service) ListMyRestaurants(ownerID string) ([]*Restaurant, error) {
	return s.repo.ListByOwner(ownerID)
}

// --------------------------------------------------
// Competitive insight (READ ONLY)
// --------------------------------------------------
func (s *Service) GetCompetitiveInsight(
	ctx context.Context,
	restaurantID int,
	userID string,
) (*CompetitiveInsight, error) {

	// 🔒 Ownership enforced here
	isOwner, err := s.repo.IsOwner(ctx, restaurantID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("unauthorized")
	}

	cost, city, cuisine, err := s.repo.GetLatestParsedCostForTwo(
		ctx,
		restaurantID,
	)
	if err != nil {
		return nil, errors.New("no parsed menu available")
	}

	snapshot, err := s.competitionRepo.GetSnapshot(ctx, city, cuisine)
	if err != nil {
		return nil, errors.New("no competitive data available")
	}

	position := determinePosition(cost, snapshot.MedianCostForTwo)

	return &CompetitiveInsight{
		RestaurantID:         restaurantID,
		City:                 city,
		CuisineType:          cuisine,
		RestaurantCostForTwo: cost,
		MarketAvg:            snapshot.AvgCostForTwo,
		MarketMedian:         snapshot.MedianCostForTwo,
		SampleSize:           snapshot.SampleSize,
		Positioning:          position,
	}, nil
}

// --------------------------------------------------
// Positioning logic
// --------------------------------------------------
func determinePosition(cost, median float64) string {
	switch {
	case cost < median*0.9:
		return "UNDER_MARKET"
	case cost > median*1.1:
		return "PREMIUM"
	default:
		return "MARKET_AVERAGE"
	}
}

