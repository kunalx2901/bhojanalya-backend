package restaurant

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
	"bhojanalya/internal/menu"
	"bhojanalya/internal/competition"
	"bhojanalya/internal/storage"
)

type Service struct {
	repo            Repository
	menuService     *menu.Service
	competitionRepo *competition.Repository
	r2              *storage.R2Client
}

func NewService(
	repo Repository,
	menuService *menu.Service,
	competitionRepo *competition.Repository,
	r2 *storage.R2Client,
) *Service {
	return &Service{
		repo:            repo,
		menuService:     menuService,
		competitionRepo: competitionRepo,
		r2:              r2,
	}
}


// --------------------------------------------------
// Create restaurant (with description + timings)
// --------------------------------------------------
func (s *Service) CreateRestaurant(
	name string,
	city string,
	cuisineType string,
	shortDescription string,
	opensAt string,
	closesAt string,
	ownerID string,
) (*Restaurant, error) {

	if name == "" || city == "" || cuisineType == "" {
		return nil, errors.New("missing required fields")
	}

	if opensAt != "" && closesAt != "" {
		oa, err1 := time.Parse("15:04", opensAt)
		ca, err2 := time.Parse("15:04", closesAt)
		if err1 != nil || err2 != nil {
			return nil, errors.New("invalid time format, expected HH:MM")
		}
		if !oa.Before(ca) {
			return nil, errors.New("opens_at must be before closes_at")
		}
	}

	restaurant := &Restaurant{
		Name:             name,
		City:             city,
		CuisineType:      cuisineType,
		ShortDescription: shortDescription,
		OpensAt:          opensAt,
		ClosesAt:         closesAt,
		OwnerID:          ownerID,
		Status:           "pending",
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
// ADMIN: List approved restaurants
// --------------------------------------------------
func (s *Service) ListApprovedRestaurants(
	ctx context.Context,
) ([]*Restaurant, error) {
	return s.repo.ListApproved(ctx)
}

// --------------------------------------------------
// ADMIN: View restaurant details
// --------------------------------------------------
func (s *Service) GetAdminRestaurantDetails(
	ctx context.Context,
	restaurantID int,
) (*AdminRestaurantDetails, error) {
	return s.repo.GetAdminDetails(ctx, restaurantID)
}

// --------------------------------------------------
// Competitive insight (READ ONLY)
// --------------------------------------------------
func (s *Service) GetCompetitiveInsight(
	ctx context.Context,
	restaurantID int,
	userID string,
) (*CompetitiveInsight, error) {

	ok, err := s.repo.IsOwner(ctx, restaurantID, userID)
	if err != nil || !ok {
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
// Upload restaurant images (STORE OBJECT KEYS ONLY)
// --------------------------------------------------
func (s *Service) UploadImages(
	ctx context.Context,
	restaurantID int,
	userID string,
	files []*multipart.FileHeader,
) error {

	ok, err := s.repo.IsOwner(ctx, restaurantID, userID)
	if err != nil || !ok {
		return errors.New("unauthorized")
	}

	var imageKeys []string

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

		_, err := storage.UploadMultipartFile(
			ctx,
			s.r2.GetClient(),
			s.r2.GetBucket(),
			key,
			file,
		)
		if err != nil {
			return err
		}

		// ✅ store OBJECT KEY, not URL
		imageKeys = append(imageKeys, key)
	}

	return s.repo.SaveRestaurantImages(ctx, restaurantID, imageKeys)
}

// --------------------------------------------------
// Preview (SIGNED URLs)
// --------------------------------------------------
func (s *Service) GetPreview(
	ctx context.Context,
	restaurantID int,
	userID string,
) (*PreviewData, error) {

	ok, err := s.repo.IsOwner(ctx, restaurantID, userID)
	if err != nil || !ok {
		return nil, errors.New("unauthorized")
	}

	hasDeal, err := s.repo.HasAnyDeal(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	if !hasDeal {
		return nil, errors.New("preview unavailable until at least one deal exists")
	}

	preview, err := s.repo.GetPreviewData(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	for i, key := range preview.Images {
		if url, err := s.r2.GetSignedURL(ctx, key, 15*time.Minute); err == nil {
			preview.Images[i] = url
		}
	}

	for i, key := range preview.MenuPDFs {
		if url, err := s.r2.GetSignedURL(ctx, key, 15*time.Minute); err == nil {
			preview.MenuPDFs[i] = url
		}
	}

	return preview, nil
}

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

// --------------------------------------------------
// ADMIN: Approve restaurant (menu + restaurant + deals)
// --------------------------------------------------
func (s *Service) ApproveRestaurant(
	ctx context.Context,
	restaurantID int,
	adminID string,
) error {
	// Delegate to menu approval logic
	return s.menuService.ApproveRestaurant(ctx, restaurantID, adminID)
}

