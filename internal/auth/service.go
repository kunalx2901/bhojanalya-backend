package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Service contains business logic.
type Service struct {
	repo UserRepository
}

// Constructor (IMPORTANT: accepts interface, not struct)
func NewService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

// Register registers a new user with hashed password.
func (s *Service) Register(name, email, password string) (*User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("missing required fields")
	}

	exists, err := s.repo.ExistsByEmail(email)
	if err != nil {
		return nil, err
	}
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
	}

	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}
