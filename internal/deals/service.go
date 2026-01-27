package deals

import (
	"context"
	"errors"

	"bhojanalya/internal/competition"
	"bhojanalya/internal/core"
)

type Service struct {
	repo              *Repository
	restaurantReader  core.RestaurantReader
	competitionRepo   *competition.Repository
}

func NewService(
	repo *Repository,
	restaurantReader core.RestaurantReader,
	competitionRepo *competition.Repository,
) *Service {
	return &Service{
		repo:             repo,
		restaurantReader: restaurantReader,
		competitionRepo:  competitionRepo,
	}
}
func (s *Service) GetDealSuggestion(
	ctx context.Context,
	restaurantID int,
	userID string,
) (*DealSuggestion, error) {

	// 🔒 Ownership check
	ok, err := s.restaurantReader.IsOwner(ctx, restaurantID, userID)
	if err != nil || !ok {
		return nil, errors.New("unauthorized")
	}

	// 📊 Restaurant pricing
	cost, city, cuisine, err :=
		s.restaurantReader.GetLatestParsedCostForTwo(ctx, restaurantID)
	if err != nil {
		return nil, errors.New("no menu data")
	}

	// 📈 Market snapshot
	snap, err := s.competitionRepo.GetSnapshot(ctx, city, cuisine)
	if err != nil {
		return nil, errors.New("no market data")
	}

	var suggestions []SuggestedDeal

	// ---------- STARTERS ----------
	starter := "starter"
	suggestions = append(suggestions,
		SuggestedDeal{
			Type:          "PERCENTAGE",
			Category:      &starter,
			Title:         "10% off on Starters",
			DiscountValue: 10,
			Reason:        "Increase early ordering and table conversion",
		},
		SuggestedDeal{
			Type:          "FLAT",
			Category:      &starter,
			Title:         "₹100 off on Starters above ₹499",
			DiscountValue: 100,
			Reason:        "Encourage higher starter basket size",
		},
	)

	// ---------- MAIN COURSE ----------
	main := "main_course"
	if cost > snap.MedianCostForTwo*1.1 {
		suggestions = append(suggestions,
			SuggestedDeal{
				Type:          "PERCENTAGE",
				Category:      &main,
				Title:         "15% off on Main Course",
				DiscountValue: 15,
				Reason:        "Your pricing is higher than competitors",
			},
			SuggestedDeal{
				Type:          "COMBO",
				Category:      &main,
				Title:         "Main Course + Starter Combo",
				DiscountValue: 0,
				Reason:        "Bundle to reduce price perception",
			},
		)
	}

	// ---------- DRINKS ----------
	drinks := "drink"
	suggestions = append(suggestions,
		SuggestedDeal{
			Type:          "PERCENTAGE",
			Category:      &drinks,
			Title:         "Happy Hour – 20% off Drinks",
			DiscountValue: 20,
			Reason:        "Increase footfall during off-peak hours",
		},
		SuggestedDeal{
			Type:          "FLAT",
			Category:      &drinks,
			Title:         "Buy 1 Get 1 on Soft Drinks",
			DiscountValue: 50,
			Reason:        "Low cost, high perceived value",
		},
	)

	// ---------- DESSERT ----------
	dessert := "dessert"
	suggestions = append(suggestions,
		SuggestedDeal{
			Type:          "PERCENTAGE",
			Category:      &dessert,
			Title:         "Free Dessert with Main Course",
			DiscountValue: 100,
			Reason:        "Improve meal completion and satisfaction",
		},
		SuggestedDeal{
			Type:          "FLAT",
			Category:      &dessert,
			Title:         "₹75 off on Desserts",
			DiscountValue: 75,
			Reason:        "Increase dessert attachment rate",
		},
	)

	// 📦 Final response
	return &DealSuggestion{
		RestaurantID:         restaurantID,
		City:                 city,
		CuisineType:          cuisine,
		Positioning:          determinePosition(cost, snap.MedianCostForTwo),
		RestaurantCostForTwo: cost,
		MarketAvg:            snap.AvgCostForTwo,
		MarketMedian:         snap.MedianCostForTwo,
		Suggestions:          suggestions,
	}, nil
}

// ---------------------------------------------
// Create Deal (Restaurant action)
// ----------------------------------------------
func (s *Service) CreateDeal(
	ctx context.Context,
	userID string,
	deal *Deal,
) error {

	// 🔒 Ownership check
	ok, err := s.restaurantReader.IsOwner(ctx, deal.RestaurantID, userID)
	if err != nil || !ok {
		return errors.New("unauthorized")
	}

	// Workflow defaults
	deal.Status = "PENDING_APPROVAL"
	deal.Suggested = false

	return s.repo.Create(ctx, deal)
}


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
