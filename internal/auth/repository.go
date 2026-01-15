package auth

// UserRepository defines the data-access contract.
// Service depends ONLY on this interface.
type UserRepository interface {
	Save(user *User) error
	ExistsByEmail(email string) (bool, error)
}
