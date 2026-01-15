package auth

type UserRepository interface {
	Save(user *User) error
	ExistsByEmail(email string) (bool, error)
	FindByEmail(email string) (*User, error)
}
