package restaurant

import (
	"testing"
	"time"
)

// MockRepository is a mock implementation of Repository interface for testing
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
	restaurant.ID = "mock-id-" + restaurant.Name
	restaurant.CreatedAt = time.Now()
	m.restaurants[restaurant.OwnerID] = append(m.restaurants[restaurant.OwnerID], restaurant)
	return nil
}

func (m *MockRepository) ListByOwner(ownerID string) ([]*Restaurant, error) {
	return m.restaurants[ownerID], nil
}

// TestCreateRestaurant_Success tests successful restaurant creation
func TestCreateRestaurant_Success(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	restaurant, err := service.CreateRestaurant("Taj Palace", "New York", "Indian", "owner-123")

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

// TestCreateRestaurant_MissingName tests restaurant creation with missing name
func TestCreateRestaurant_MissingName(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	_, err := service.CreateRestaurant("", "New York", "Indian", "owner-123")

	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}

	if err.Error() != "missing required fields" {
		t.Errorf("expected 'missing required fields', got '%s'", err.Error())
	}
}

// TestCreateRestaurant_MissingCity tests restaurant creation with missing city
func TestCreateRestaurant_MissingCity(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	_, err := service.CreateRestaurant("Taj Palace", "", "Indian", "owner-123")

	if err == nil {
		t.Fatal("expected error for missing city, got nil")
	}

	if err.Error() != "missing required fields" {
		t.Errorf("expected 'missing required fields', got '%s'", err.Error())
	}
}

// TestCreateRestaurant_MissingCuisineType tests restaurant creation with missing cuisine type
func TestCreateRestaurant_MissingCuisineType(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	_, err := service.CreateRestaurant("Taj Palace", "New York", "", "owner-123")

	if err == nil {
		t.Fatal("expected error for missing cuisine type, got nil")
	}

	if err.Error() != "missing required fields" {
		t.Errorf("expected 'missing required fields', got '%s'", err.Error())
	}
}

// TestListMyRestaurants_Success tests successful retrieval of user's restaurants
func TestListMyRestaurants_Success(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	// Create multiple restaurants
	service.CreateRestaurant("Taj Palace", "New York", "Indian", "owner-123")
	service.CreateRestaurant("Dragon Court", "New York", "Chinese", "owner-123")
	service.CreateRestaurant("Pasta Paradise", "Boston", "Italian", "owner-456")

	// List restaurants for owner-123
	restaurants, err := service.ListMyRestaurants("owner-123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(restaurants) != 2 {
		t.Errorf("expected 2 restaurants, got %d", len(restaurants))
	}

	if restaurants[0].Name != "Taj Palace" {
		t.Errorf("expected first restaurant 'Taj Palace', got '%s'", restaurants[0].Name)
	}

	if restaurants[1].Name != "Dragon Court" {
		t.Errorf("expected second restaurant 'Dragon Court', got '%s'", restaurants[1].Name)
	}
}

// TestListMyRestaurants_Empty tests retrieval when user has no restaurants
func TestListMyRestaurants_Empty(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	restaurants, err := service.ListMyRestaurants("owner-no-restaurants")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if restaurants != nil && len(restaurants) != 0 {
		t.Errorf("expected empty list, got %d restaurants", len(restaurants))
	}
}
