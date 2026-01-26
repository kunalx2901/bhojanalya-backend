package deals

import (
	"context"
	"errors"

	"bhojanalya/internal/competition"
	"bhojanalya/internal/core"
)

type Service struct {
	restaurantReader core.RestaurantReader
	competitionRepo  *competition.Repository
}

func NewService(
	restaurantReader core.RestaurantReader,
	competitionRepo *competition.Repository,
) *Service {
	return &Service{
		restaurantReader: restaurantReader,
		competitionRepo:  competitionRepo,
	}
}

func (s *Service) GetDealSuggestion(
	ctx context.Context,
	restaurantID int,
	userID string,
) (*DealSuggestion, error) {

	ok, err := s.restaurantReader.IsOwner(ctx, restaurantID, userID)
	if err != nil || !ok {
		return nil, errors.New("unauthorized")
	}

	cost, city, cuisine, err :=
		s.restaurantReader.GetLatestParsedCostForTwo(ctx, restaurantID)
	if err != nil {
		return nil, errors.New("no menu data")
	}

	snap, err := s.competitionRepo.GetSnapshot(ctx, city, cuisine)
	if err != nil {
		return nil, errors.New("no market data")
	}

	action := "HAPPY_HOUR"
	discount := 10
	reason := "Balanced pricing — increase footfall"

	if cost > snap.MedianCostForTwo*1.1 {
		action = "DISCOUNT_MAIN_COURSE"
		discount = 15
		reason = "Priced higher than competitors"
	} else if cost < snap.MedianCostForTwo*0.9 {
		action = "PREMIUM_COMBO"
		discount = 0
		reason = "You are priced competitively — upsell combos"
	}

	return &DealSuggestion{
		RestaurantID:         restaurantID,
		City:                 city,
		CuisineType:          cuisine,
		RestaurantCostForTwo: cost,
		MarketAvg:            snap.AvgCostForTwo,
		MarketMedian:         snap.MedianCostForTwo,
		Positioning:          action,
		SuggestedAction:      action,
		SuggestedDiscount:    discount,
		Reason:               reason,
	}, nil
}
