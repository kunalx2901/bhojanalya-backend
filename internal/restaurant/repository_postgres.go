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

// --------------------------------------------------
// Create a new restaurant
// --------------------------------------------------
func (r *PostgresRepository) Create(restaurant *Restaurant) error {
	query := `
		INSERT INTO restaurants (
			name,
			city,
			cuisine_type,
			owner_id,
			status
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	return r.db.QueryRow(
		context.Background(),
		query,
		restaurant.Name,
		restaurant.City,
		restaurant.CuisineType,
		restaurant.OwnerID,
		restaurant.Status,
	).Scan(&restaurant.ID, &restaurant.CreatedAt)
}

// --------------------------------------------------
// List restaurants owned by a user
// --------------------------------------------------
func (r *PostgresRepository) ListByOwner(ownerID string) ([]*Restaurant, error) {
	query := `
		SELECT
			id,
			name,
			city,
			cuisine_type,
			owner_id,
			status,
			created_at
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

// --------------------------------------------------
// Get latest PARSED cost-for-two for restaurant
// Used for competitive insights (READ-ONLY)
// --------------------------------------------------
func (r *PostgresRepository) GetLatestParsedCostForTwo(
	ctx context.Context,
	restaurantID int,
) (float64, string, string, error) {

	var cost float64
	var city string
	var cuisine string

	err := r.db.QueryRow(ctx, `
		SELECT
			(mu.parsed_data->'cost_for_two'->'calculation'->>'total_cost_for_two')::numeric,
			r.city,
			r.cuisine_type
		FROM menu_uploads mu
		JOIN restaurants r
		  ON r.id = mu.restaurant_id
		WHERE
			mu.restaurant_id = $1
			AND mu.status = 'PARSED'
			AND mu.parsed_data IS NOT NULL
			AND mu.parsed_data->'cost_for_two'->'calculation'->>'total_cost_for_two' IS NOT NULL
		ORDER BY mu.updated_at DESC
		LIMIT 1
	`, restaurantID).Scan(&cost, &city, &cuisine)

	return cost, city, cuisine, err
}

// --------------------------------------------------
// Ownership check (SECURITY)
// --------------------------------------------------
func (r *PostgresRepository) IsOwner(
	ctx context.Context,
	restaurantID int,
	userID string,
) (bool, error) {

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM restaurants
			WHERE id = $1
			  AND owner_id = $2
		)
	`, restaurantID, userID).Scan(&exists)

	return exists, err
}
