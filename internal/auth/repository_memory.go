package auth

import "errors"

type InMemoryUserRepository struct {
	users map[string]*User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*User),
	}
}

func (r *InMemoryUserRepository) Save(user *User) error {
	r.users[user.Email] = user
	return nil
}

func (r *InMemoryUserRepository) ExistsByEmail(email string) (bool, error) {
	_, exists := r.users[email]
	return exists, nil
}

func (r *InMemoryUserRepository) FindByEmail(email string) (*User, error) {
	user, ok := r.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}
