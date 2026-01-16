package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(user *User) error {
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(context.Background(), query,
		user.Name, user.Email, user.Password,
	)
	return err
}

func (r *PostgresUserRepository) ExistsByEmail(email string) (bool, error) {
	query := `SELECT 1 FROM users WHERE email=$1 LIMIT 1`
	row := r.db.QueryRow(context.Background(), query, email)

	var exists int
	err := row.Scan(&exists)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (r *PostgresUserRepository) FindByEmail(email string) (*User, error) {
	query := `
		SELECT name, email, password
		FROM users WHERE email=$1
	`
	row := r.db.QueryRow(context.Background(), query, email)

	user := &User{}
	if err := row.Scan(&user.Name, &user.Email, &user.Password); err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
