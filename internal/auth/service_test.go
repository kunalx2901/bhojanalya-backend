package auth

import "testing"

func TestPasswordIsHashedBeforeSaving(t *testing.T) {
	repo := NewInMemoryUserRepository()
	service := NewService(repo)

	password := "Password@123"

	_, err := service.Register("Test User", "test@example.com", password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user := repo.users["test@example.com"]
	if user == nil {
		t.Fatalf("user not found")
	}

	if user.Password == password {
		t.Fatalf("password was stored in plain text")
	}
}
