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


// save_ocr_text saves the extracted OCR text for a given menu upload.
func (r *PostgresRepository) SaveMenuItems(
	menuUploadID int,
	restaurantID int,
	items []MenuItem,
) error {

	if len(items) == 0 {
		return nil
	}

	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO menu_items
			(menu_upload_id, restaurant_id, name, category, price, confidence)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			menuUploadID,
			restaurantID,
			item.Name,
			item.Category,
			item.Price,
			item.Confidence,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
