package auth

import "context"

type UserRepository interface {
	Save(user *User) error
	ExistsByEmail(email string) (bool, error)
	FindByEmail(email string) (*User, error)
	// Add these two:
	GetOnboardingStatus(ctx context.Context, userID string) (string, error)
	UpdateOnboardingStatus(ctx context.Context, userID string, status string) error
}
