package restaurant

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// create a new restaurant record in the database
func (r *PostgresRepository) Create(restaurant *Restaurant) error {
	query := `
		INSERT INTO restaurants (name, city, cuisine_type, owner_id, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(context.Background(),
		query,
		restaurant.Name,
		restaurant.City,
		restaurant.CuisineType,
		restaurant.OwnerID,
		restaurant.Status,
	).Scan(&restaurant.ID, &restaurant.CreatedAt)

	return err
}


// list restaurants by owner ID
func (r *PostgresRepository) ListByOwner(ownerID string) ([]*Restaurant, error) {
	query := `
		SELECT id, name, city, cuisine_type, owner_id, status, created_at
		FROM restaurants
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(context.Background(), query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var restaurants []*Restaurant

	for rows.Next() {
		var res Restaurant
		if err := rows.Scan(
			&res.ID,
			&res.Name,
			&res.City,
			&res.CuisineType,
			&res.OwnerID,
			&res.Status,
			&res.CreatedAt,
		); err != nil {
			return nil, err
		}
		restaurants = append(restaurants, &res)
	}

	return restaurants, nil
}

