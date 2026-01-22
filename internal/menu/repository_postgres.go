package menu

import (
	"context"
	"encoding/json"
	"errors"

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

// SaveMenuItems saves individual menu items (legacy support)
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

// SaveParsedMenu saves the canonical parsed JSON + cost-for-two
func (r *PostgresRepository) SaveParsedMenu(
	menuUploadID int,
	doc map[string]interface{},
) error {

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	cmd, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET parsed_data = $1,
		    status = 'PARSED'
		WHERE id = $2
		`,
		data,
		menuUploadID,
	)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("no menu_upload row updated")
	}

	return nil
}
