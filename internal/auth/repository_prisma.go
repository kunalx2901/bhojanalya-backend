package auth

import (
	"context"
	"bhojanalya/prisma/db" // Adjust the path based on your module name in go.mod
)

type PrismaUserRepository struct {
	client *db.PrismaClient
}

func NewPrismaUserRepository(client *db.PrismaClient) *PrismaUserRepository {
	return &PrismaUserRepository{client: client}
}

// Save stores a user in the database
func (r *PrismaUserRepository) Save(user *User) error {
	ctx := context.Background()
	_, err := r.client.User.CreateOne(
		db.User.Name.Set(user.Name),
		db.User.Email.Set(user.Email),
		db.User.Password.Set(user.Password),
	).Exec(ctx)
	return err
}

// ExistsByEmail checks if a user exists in the database
func (r *PrismaUserRepository) ExistsByEmail(email string) (bool, error) {
	ctx := context.Background()
	user, err := r.client.User.FindUnique(
		db.User.Email.Equals(email),
	).Exec(ctx)

	if err != nil {
		// Prisma returns an error if the record is not found
		return false, nil 
	}
	return user != nil, nil
}