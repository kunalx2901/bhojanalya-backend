package competition

import (
	"context"
	"log"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db   *pgxpool.Pool
	repo *Repository
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:   db,
		repo: NewRepository(db),
	}
}

// Recompute snapshot for a city + cuisine
func (s *Service) RecomputeSnapshot(
	ctx context.Context,
	city string,
	cuisine string,
) error {

	rows, err := s.db.Query(ctx, `
		SELECT
			(parsed_data->'cost_for_two'->'calculation'->>'total_cost_for_two')::numeric
		FROM menu_uploads mu
		JOIN restaurants r
		  ON mu.restaurant_id = r.id
		WHERE
			mu.status = 'PARSED'
			AND mu.parsed_data IS NOT NULL
			AND r.city = $1
			AND r.cuisine_type = $2
	`, city, cuisine)
	if err != nil {
		return err
	}
	defer rows.Close()

	var values []float64

	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err == nil {
			values = append(values, v)
		}
	}

	// ✅ Allow at least 1 sample (DEV / TEST)
	if len(values) == 0 {
		log.Printf(
			"[COMPETITION] No data for %s / %s",
			city, cuisine,
		)
		return nil
	}

	sort.Float64s(values)

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	avg := sum / float64(len(values))

	median := values[len(values)/2]
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	}

	log.Printf(
		"[COMPETITION] %s / %s → avg=%.2f median=%.2f samples=%d",
		city, cuisine, avg, median, len(values),
	)

	return s.repo.UpsertSnapshot(ctx, Snapshot{
		City:             city,
		CuisineType:      cuisine,
		AvgCostForTwo:    avg,
		MedianCostForTwo: median,
		SampleSize:       len(values),
	})
}

// Read-only fetch for API
func (s *Service) GetSnapshot(
	ctx context.Context,
	city string,
	cuisine string,
) (*Snapshot, error) {
	return s.repo.GetSnapshot(ctx, city, cuisine)
}
