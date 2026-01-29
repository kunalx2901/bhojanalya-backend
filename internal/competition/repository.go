package competition

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

// Insert or update snapshot for (city, cuisine_type)
func (r *Repository) UpsertSnapshot(
	ctx context.Context,
	s Snapshot,
) error {

	_, err := r.db.Exec(ctx, `
		INSERT INTO competitive_snapshots (
			city,
			cuisine_type,
			avg_cost_for_two,
			median_cost_for_two,
			sample_size
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (city, cuisine_type)
		DO UPDATE SET
			avg_cost_for_two = EXCLUDED.avg_cost_for_two,
			median_cost_for_two = EXCLUDED.median_cost_for_two,
			sample_size = EXCLUDED.sample_size,
			updated_at = now()
	`,
		s.City,
		s.CuisineType,
		s.AvgCostForTwo,
		s.MedianCostForTwo,
		s.SampleSize,
	)

	return err
}

// Fetch snapshot for API
func (r *Repository) GetSnapshot(
	ctx context.Context,
	city string,
	cuisine string,
) (*Snapshot, error) {

	var s Snapshot
	err := r.db.QueryRow(ctx, `
		SELECT
			id,
			city,
			cuisine_type,
			avg_cost_for_two,
			median_cost_for_two,
			sample_size,
			created_at,
			updated_at
		FROM competitive_snapshots
		WHERE city = $1 AND cuisine_type = $2
	`, city, cuisine).Scan(
		&s.ID,
		&s.City,
		&s.CuisineType,
		&s.AvgCostForTwo,
		&s.MedianCostForTwo,
		&s.SampleSize,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
