package auth

import (
	"context"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	AdminEmail            = "admin@bhojanalya.com"
	AdminPassword         = "Bhojanalya@12345"
)

type Service struct {
	repo UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

// REGISTER
func (s *Service) Register(name, email, password string) (*User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("missing required fields")
	}

	exists, _ := s.repo.ExistsByEmail(email)
	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user := &User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Role:     string(RoleRestaurant),
	}

	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

// LOGIN
func (s *Service) Login(email, password string) (*User, error) {
	log.Printf("Login attempt for email: %s", email)

	// Check if it's the admin credentials
	if email == AdminEmail && password == AdminPassword {
		log.Printf("Admin login successful for email: %s", email)

		// Check if admin already exists in database
		adminUser, err := s.repo.FindByEmail(email)
		if err != nil {
			// Admin doesn't exist, create them with ADMIN role
			hashedPassword, err := bcrypt.GenerateFromPassword(
				[]byte(password),
				bcrypt.DefaultCost,
			)
			if err != nil {
				return nil, err
			}

			adminUser := &User{
				Name:     "Bhojanalya Admin",
				Email:    email,
				Password: string(hashedPassword),
				Role:     string(RoleAdmin), // Stored as ADMIN role in database
			}

			if err := s.repo.Save(adminUser); err != nil {
				log.Printf("Error saving admin user: %v", err)
				return nil, err
			}

			log.Printf("Admin user created in database with ADMIN role")
			return adminUser, nil
		}

		// Admin exists, verify password and return
		err = bcrypt.CompareHashAndPassword(
			[]byte(adminUser.Password),
			[]byte(password),
		)
		if err != nil {
			log.Printf("Admin password mismatch for email: %s", email)
			return nil, ErrInvalidCredentials
		}

		return adminUser, nil
	}

	user, err := s.repo.FindByEmail(email)
	if err != nil {
		log.Printf("User not found: %s", email)
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		log.Printf("Password mismatch for email: %s, error: %v", email, err)
		return nil, ErrInvalidCredentials
	}

	log.Printf("Login successful for email: %s with role: %s", email, user.Role)
	return user, nil
}
func (s *Service) GetOnboardingStatus(ctx context.Context, userID string) (string, error) {
	return s.repo.GetOnboardingStatus(ctx, userID)
}

func (s *Service) UpdateOnboardingStatus(ctx context.Context, userID string, status string) error {
	return s.repo.UpdateOnboardingStatus(ctx, userID, status)
}
