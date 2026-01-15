package auth

// InMemoryUserRepository is a simple in-memory implementation
// of UserRepository used for tests and local development.
type InMemoryUserRepository struct {
	users map[string]*User
}

// Constructor
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*User),
	}
}

// Save stores a user in memory.
func (r *InMemoryUserRepository) Save(user *User) error {
	r.users[user.Email] = user
	return nil
}

// ExistsByEmail checks if a user exists by email.
func (r *InMemoryUserRepository) ExistsByEmail(email string) (bool, error) {
	_, exists := r.users[email]
	return exists, nil
}
