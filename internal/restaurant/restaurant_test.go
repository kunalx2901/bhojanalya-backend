package restaurant

import (
	"context"
	"strconv"
	"testing"
	"time"

	"bhojanalya/internal/competition"
)

// --------------------------------------------------
// Mock Repository
// --------------------------------------------------

type MockRepository struct {
	restaurants map[string][]*Restaurant
	createErr   error
	nextID      int
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		restaurants: make(map[string][]*Restaurant),
		nextID:      1,
	}
}

func (m *MockRepository) Create(restaurant *Restaurant) error {
	if m.createErr != nil {
		return m.createErr
	}

	restaurant.ID = strconv.Itoa(m.nextID)
	m.nextID++
	restaurant.CreatedAt = time.Now()

	m.restaurants[restaurant.OwnerID] = append(
		m.restaurants[restaurant.OwnerID],
		restaurant,
	)
	return nil
}

func (m *MockRepository) ListByOwner(ownerID string) ([]*Restaurant, error) {
	return m.restaurants[ownerID], nil
}

// --------------------------------------------------
// REQUIRED BY Repository INTERFACE (NO-OP)
// --------------------------------------------------

func (m *MockRepository) IsOwner(
	ctx context.Context,
	restaurantID int,
	userID string,
) (bool, error) {
	return true, nil
}

func (m *MockRepository) GetLatestParsedCostForTwo(
	ctx context.Context,
	restaurantID int,
) (float64, string, string, error) {
	return 0, "", "", nil
}

func (m *MockRepository) HasAnyDeal(
	ctx context.Context,
	restaurantID int,
) (bool, error) {
	return true, nil
}

func (m *MockRepository) GetPreviewData(
	ctx context.Context,
	restaurantID int,
) (*PreviewData, error) {
	return &PreviewData{
		Name:        "Mock Restaurant",
		City:        "Mock City",
		CuisineType: "Mock Cuisine",
		CostForTwo:  500,
		Images:      []string{},
	}, nil
}

func (m *MockRepository) SaveRestaurantImages(
	ctx context.Context,
	restaurantID int,
	images []string,
) error {
	return nil
}

func (m *MockRepository) GetRestaurantImages(
	ctx context.Context,
	restaurantID int,
) ([]string, error) {
	return []string{}, nil
}

// 🔥 ADMIN METHODS (NEW, REQUIRED)
func (m *MockRepository) ListApproved(
	ctx context.Context,
) ([]*Restaurant, error) {
	return []*Restaurant{}, nil
}

func (m *MockRepository) GetAdminDetails(
	ctx context.Context,
	restaurantID int,
) (*AdminRestaurantDetails, error) {
	return &AdminRestaurantDetails{
		ID:          restaurantID,
		Email:       "owner@test.com",
		OwnerName:   "Test Owner",
		CuisineType:"Indian",
		City:        "Test City",
	}, nil
}

// --------------------------------------------------
// TESTS
// --------------------------------------------------

func TestCreateRestaurant_Success(t *testing.T) {
	mockRepo := NewMockRepository()

	service := NewService(
		mockRepo,
		&competition.Repository{},
		nil,
	)

	restaurant, err := service.CreateRestaurant(
		"Taj Palace",
		"New York",
		"Indian",
		"Luxury Indian dining",
		"10:00",
		"23:00",
		"owner-123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if restaurant.ID == "" {
		t.Errorf("expected ID to be set")
	}

	if restaurant.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", restaurant.Status)
	}
}

func TestCreateRestaurant_MissingFields(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo, &competition.Repository{}, nil)

	_, err := service.CreateRestaurant(
		"",
		"NY",
		"Indian",
		"",
		"",
		"",
		"owner",
	)
	if err == nil {
		t.Fatal("expected error for missing fields")
	}
}

func TestListMyRestaurants_Success(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo, &competition.Repository{}, nil)

	service.CreateRestaurant("Taj Palace", "NY", "Indian", "", "", "", "owner-123")
	service.CreateRestaurant("Dragon Court", "NY", "Chinese", "", "", "", "owner-123")
	service.CreateRestaurant("Pasta House", "Boston", "Italian", "", "", "", "owner-456")

	restaurants, err := service.ListMyRestaurants("owner-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(restaurants) != 2 {
		t.Fatalf("expected 2 restaurants, got %d", len(restaurants))
	}
}

func TestListMyRestaurants_Empty(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo, &competition.Repository{}, nil)

	restaurants, err := service.ListMyRestaurants("no-restaurants")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(restaurants) != 0 {
		t.Errorf("expected empty list, got %d", len(restaurants))
	}
}
