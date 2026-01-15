package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
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
	}

	if err := s.repo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

// LOGIN
func (s *Service) Login(email, password string) (*User, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
