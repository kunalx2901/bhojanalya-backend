package restaurant

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"bhojanalya/internal/competition"
	"bhojanalya/internal/storage"
)

type Service struct {
	repo            Repository
	competitionRepo *competition.Repository
	r2              *storage.R2Client
}

func NewService(
	repo Repository,
	competitionRepo *competition.Repository,
	r2 *storage.R2Client,
) *Service {
	return &Service{
		repo:            repo,
		competitionRepo: competitionRepo,
		r2:              r2,
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

	isOwner, err := s.repo.IsOwner(ctx, restaurantID, userID)
	if err != nil || !isOwner {
		return nil, errors.New("unauthorized")
	}

	cost, city, cuisine, err :=
		s.repo.GetLatestParsedCostForTwo(ctx, restaurantID)
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
// Upload restaurant images
// --------------------------------------------------
func (s *Service) UploadImages(
	ctx context.Context,
	restaurantID int,
	userID string,
	files []*multipart.FileHeader,
) error {

	// 🔒 Ownership check
	ok, err := s.repo.IsOwner(ctx, restaurantID, userID)
	if err != nil || !ok {
		return errors.New("unauthorized")
	}

	var imageURLs []string

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			return errors.New("only jpg, jpeg, png images allowed")
		}

		key := fmt.Sprintf(
			"restaurants/%d/%s",
			restaurantID,
			file.Filename,
		)

		url, err := storage.UploadMultipartFile(
			ctx,
			s.r2.GetClient(),
			s.r2.GetBucket(),
			key,
			file,
		)
		if err != nil {
			return err
		}

		imageURLs = append(imageURLs, url)
	}

	return s.repo.SaveRestaurantImages(ctx, restaurantID, imageURLs)
}

// --------------------------------------------------
// Preview (only after deal creation)
// --------------------------------------------------
func (s *Service) GetPreview(
	ctx context.Context,
	restaurantID int,
	userID string,
) (*PreviewData, error) {

	// 🔒 Ownership
	ok, err := s.repo.IsOwner(ctx, restaurantID, userID)
	if err != nil || !ok {
		return nil, errors.New("unauthorized")
	}

	// 🔒 Must have at least one deal
	hasDeal, err := s.repo.HasAnyDeal(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	if !hasDeal {
		return nil, errors.New("preview unavailable until at least one deal exists")
	}

	return s.repo.GetPreviewData(ctx, restaurantID)
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
