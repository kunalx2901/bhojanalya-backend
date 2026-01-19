package menu

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

// to save a menu record in the database
func (r *PostgresRepository) Create(menu *Menu) error {
	query := `
		INSERT INTO menus (restaurant_id, file_path)
		VALUES ($1, $2)
	`
	_, err := r.db.Exec(context.Background(), query, menu.RestaurantID, menu.FilePath)
	return err
}

// to find a menu by restaurant ID
func (r *PostgresRepository) FindByRestaurant(restaurantID string) (*Menu, error) {
	query := `
		SELECT id, restaurant_id, file_path, uploaded_at
		FROM menus
		WHERE restaurant_id = $1
		LIMIT 1
	`

	var m Menu
	err := r.db.QueryRow(context.Background(), query, restaurantID).Scan(
		&m.ID,
		&m.RestaurantID,
		&m.FilePath,
		&m.UploadedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

