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
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		restaurants: make(map[string][]*Restaurant),
	}
}

func (m *MockRepository) Create(restaurant *Restaurant) error {
	if m.createErr != nil {
		return m.createErr
	}

	restaurant.ID = strconv.Itoa(len(m.restaurants) + 1)
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
// REQUIRED BY INTERFACE (NO-OP IMPLEMENTATIONS)
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

// --------------------------------------------------
// TESTS
// --------------------------------------------------

func TestCreateRestaurant_Success(t *testing.T) {
	mockRepo := NewMockRepository()

	service := NewService(
		mockRepo,
		&competition.Repository{}, // not used in these tests
	)

	restaurant, err := service.CreateRestaurant(
		"Taj Palace",
		"New York",
		"Indian",
		"owner-123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if restaurant.Name != "Taj Palace" {
		t.Errorf("expected name 'Taj Palace', got '%s'", restaurant.Name)
	}

	if restaurant.City != "New York" {
		t.Errorf("expected city 'New York', got '%s'", restaurant.City)
	}

	if restaurant.CuisineType != "Indian" {
		t.Errorf("expected cuisine 'Indian', got '%s'", restaurant.CuisineType)
	}

	if restaurant.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", restaurant.Status)
	}

	if restaurant.OwnerID != "owner-123" {
		t.Errorf("expected owner ID 'owner-123', got '%s'", restaurant.OwnerID)
	}
}

func TestCreateRestaurant_MissingFields(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo, &competition.Repository{})

	_, err := service.CreateRestaurant("", "NY", "Indian", "owner")
	if err == nil {
		t.Fatal("expected error for missing fields")
	}
}

func TestListMyRestaurants_Success(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo, &competition.Repository{})

	service.CreateRestaurant("Taj Palace", "NY", "Indian", "owner-123")
	service.CreateRestaurant("Dragon Court", "NY", "Chinese", "owner-123")
	service.CreateRestaurant("Pasta House", "Boston", "Italian", "owner-456")

	restaurants, err := service.ListMyRestaurants("owner-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(restaurants) != 2 {
		t.Fatalf("expected 2 restaurants, got %d", len(restaurants))
	}

	if restaurants[0].Name != "Taj Palace" {
		t.Errorf("expected 'Taj Palace', got '%s'", restaurants[0].Name)
	}
}

func TestListMyRestaurants_Empty(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo, &competition.Repository{})

	restaurants, err := service.ListMyRestaurants("no-restaurants")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(restaurants) != 0 {
		t.Errorf("expected empty list, got %d", len(restaurants))
	}
}
