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
			status,
			short_description,
			opens_at,
			closes_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
		restaurant.ShortDescription,
		restaurant.OpensAt,
		restaurant.ClosesAt,
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
			short_description,
			opens_at,
			closes_at,
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
			&res.ShortDescription,
			&res.OpensAt,
			&res.ClosesAt,
			&res.CreatedAt,
		); err != nil {
			return nil, err
		}
		restaurants = append(restaurants, &res)
	}

	return restaurants, nil
}

// --------------------------------------------------
// ADMIN: List approved restaurants
// --------------------------------------------------
func (r *PostgresRepository) ListApproved(
	ctx context.Context,
) ([]*Restaurant, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			name,
			city,
			cuisine_type,
			status,
			short_description,
			opens_at,
			closes_at,
			created_at
		FROM restaurants
		WHERE status = 'APPROVED'
		ORDER BY created_at DESC
	`)
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
			&res.Status,
			&res.ShortDescription,
			&res.OpensAt,
			&res.ClosesAt,
			&res.CreatedAt,
		); err != nil {
			return nil, err
		}
		restaurants = append(restaurants, &res)
	}

	return restaurants, nil
}

// --------------------------------------------------
// ADMIN: View restaurant details
// --------------------------------------------------
func (r *PostgresRepository) GetAdminDetails(
	ctx context.Context,
	restaurantID int,
) (*AdminRestaurantDetails, error) {

	var d AdminRestaurantDetails

	err := r.db.QueryRow(ctx, `
		SELECT
			r.id,
			u.email,
			u.name,
			r.cuisine_type,
			r.city,
			r.short_description,
			r.opens_at,
			r.closes_at
		FROM restaurants r
		JOIN users u ON u.id = r.owner_id
		WHERE r.id = $1
	`, restaurantID).Scan(
		&d.ID,
		&d.Email,
		&d.OwnerName,
		&d.CuisineType,
		&d.City,
		&d.ShortDescription,
		&d.OpensAt,
		&d.ClosesAt,
	)

	return &d, err
}

// --------------------------------------------------
// Competitive insight (READ-ONLY)
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
		LIMIT 1
	`, restaurantID).Scan(&cost, &city, &cuisine)

	return cost, city, cuisine, err
}

// --------------------------------------------------
// Ownership check
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

// --------------------------------------------------
// Preview support
// --------------------------------------------------
func (r *PostgresRepository) GetPreviewData(
	ctx context.Context,
	restaurantID int,
) (*PreviewData, error) {

	var p PreviewData
	p.ID = restaurantID

	// Core
	err := r.db.QueryRow(ctx, `
		SELECT
			name,
			city,
			cuisine_type,
			short_description,
			opens_at,
			closes_at
		FROM restaurants
		WHERE id = $1
	`, restaurantID).Scan(
		&p.Name,
		&p.City,
		&p.CuisineType,
		&p.ShortDescription,
		&p.OpensAt,
		&p.ClosesAt,
	)
	if err != nil {
		return nil, err
	}

	// Cost for two
	_ = r.db.QueryRow(ctx, `
		SELECT
			(parsed_data->'cost_for_two'->'calculation'->>'total_cost_for_two')::numeric
		FROM menu_uploads
		WHERE restaurant_id = $1
		  AND status = 'PARSED'
		LIMIT 1
	`, restaurantID).Scan(&p.CostForTwo)

	// Images
	imgRows, _ := r.db.Query(ctx, `
		SELECT image_url
		FROM restaurant_images
		WHERE restaurant_id = $1
		ORDER BY created_at
	`, restaurantID)
	defer imgRows.Close()

	for imgRows.Next() {
		var url string
		imgRows.Scan(&url)
		p.Images = append(p.Images, url)
	}

	// Menu PDFs (single menu guaranteed)
	pdfRows, _ := r.db.Query(ctx, `
		SELECT image_url
		FROM menu_uploads
		WHERE restaurant_id = $1
	`, restaurantID)
	defer pdfRows.Close()

	for pdfRows.Next() {
		var pdf string
		pdfRows.Scan(&pdf)
		p.MenuPDFs = append(p.MenuPDFs, pdf)
	}

	// Deals
	dealRows, _ := r.db.Query(ctx, `
		SELECT id, title, type, category, discount_value, status
		FROM deals
		WHERE restaurant_id = $1
		ORDER BY created_at DESC
	`, restaurantID)
	defer dealRows.Close()

	for dealRows.Next() {
		var d PreviewDeal
		dealRows.Scan(
			&d.ID,
			&d.Title,
			&d.Type,
			&d.Category,
			&d.DiscountValue,
			&d.Status,
		)
		p.Deals = append(p.Deals, d)
	}

	return &p, nil
}

// --------------------------------------------------
// Deals helper
// --------------------------------------------------
func (r *PostgresRepository) HasAnyDeal(
	ctx context.Context,
	restaurantID int,
) (bool, error) {

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM deals
			WHERE restaurant_id = $1
		)
	`, restaurantID).Scan(&exists)

	return exists, err
}

// --------------------------------------------------
// Images
// --------------------------------------------------
func (r *PostgresRepository) SaveRestaurantImages(
	ctx context.Context,
	restaurantID int,
	images []string,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, url := range images {
		_, err := tx.Exec(ctx, `
			INSERT INTO restaurant_images (restaurant_id, image_url)
			VALUES ($1, $2)
		`, restaurantID, url)

		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetRestaurantImages(
	ctx context.Context,
	restaurantID int,
) ([]string, error) {

	rows, err := r.db.Query(ctx, `
		SELECT image_url
		FROM restaurant_images
		WHERE restaurant_id = $1
		ORDER BY created_at
	`, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		images = append(images, url)
	}

	return images, nil
}
