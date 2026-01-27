package deals

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// --------------------------------------------------
// Create Deal
// --------------------------------------------------
func (r *Repository) Create(
	ctx context.Context,
	deal *Deal,
) error {

	return r.db.QueryRow(ctx, `
		INSERT INTO deals (
			restaurant_id,
			type,
			title,
			description,
			category,
			discount_value,
			original_price,
			final_price,
			status,
			suggested
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at
	`,
		deal.RestaurantID,
		deal.Type,
		deal.Title,
		deal.Description,
		deal.Category,
		deal.DiscountValue,
		deal.OriginalPrice,
		deal.FinalPrice,
		deal.Status,
		deal.Suggested,
	).Scan(
		&deal.ID,
		&deal.CreatedAt,
		&deal.UpdatedAt,
	)
}

// --------------------------------------------------
// List Deals by Restaurant
// --------------------------------------------------
func (r *Repository) ListByRestaurant(
	ctx context.Context,
	restaurantID int,
) ([]*Deal, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			restaurant_id,
			type,
			title,
			description,
			category,
			discount_value,
			original_price,
			final_price,
			status,
			suggested,
			created_at,
			updated_at
		FROM deals
		WHERE restaurant_id = $1
		ORDER BY created_at DESC
	`, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deals []*Deal

	for rows.Next() {
		var d Deal
		if err := rows.Scan(
			&d.ID,
			&d.RestaurantID,
			&d.Type,
			&d.Title,
			&d.Description,
			&d.Category,
			&d.DiscountValue,
			&d.OriginalPrice,
			&d.FinalPrice,
			&d.Status,
			&d.Suggested,
			&d.CreatedAt,
			&d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deals = append(deals, &d)
	}

	return deals, nil
}
