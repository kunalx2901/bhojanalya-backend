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

func (r *PostgresRepository) GetByID(ctx context.Context, id int) (*MenuUpload, error) {
	var upload MenuUpload
	err := r.db.QueryRow(ctx, `
		SELECT id, restaurant_id, image_url, raw_text, structured_data, status, ocr_error, created_at, updated_at
		FROM menu_uploads
		WHERE id = $1
	`, id).Scan(
		&upload.ID,
		&upload.RestaurantID,
		&upload.ImageURL,
		&upload.RawText,
		&upload.StructuredData,
		&upload.Status,
		&upload.OCRError,
		&upload.CreatedAt,
		&upload.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &upload, nil
}
