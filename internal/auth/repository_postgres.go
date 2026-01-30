package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(user *User) error {
	// Generate UUID if not already set
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	query := `
		INSERT INTO users (id, name, email, password, role)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(context.Background(), query,
		user.ID, user.Name, user.Email, user.Password, user.Role,
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
		SELECT id, name, email, password, role
		FROM users WHERE email=$1
	`
	row := r.db.QueryRow(context.Background(), query, email)

	user := &User{}
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role); err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// --------------------------------------------------
// Onboarding Status
// --------------------------------------------------

func (r *PostgresUserRepository) GetOnboardingStatus(
	ctx context.Context,
	userID string,
) (string, error) {

	var status *string

	err := r.db.QueryRow(ctx, `
		SELECT onboarding_status
		FROM users
		WHERE id = $1
	`, userID).Scan(&status)

	if err != nil {
		return "", err
	}

	if status == nil {
		return "NULL", nil
	}

	return *status, nil
}

func (r *PostgresUserRepository) UpdateOnboardingStatus(
	ctx context.Context,
	userID string,
	status string,
) error {

	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET onboarding_status = $1
		WHERE id = $2
	`, status, userID)

	return err
}
