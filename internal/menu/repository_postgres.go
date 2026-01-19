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

func (r *PostgresRepository) CreateUpload(
	restaurantID int,
	imageURL string,
	filename string,
) (int, error) {

	var id int
	err := r.db.QueryRow(context.Background(), `
		INSERT INTO menu_uploads (restaurant_id, image_url, original_filename)
		VALUES ($1, $2, $3)
		RETURNING id
	`, restaurantID, imageURL, filename).Scan(&id)

	return id, err
}
