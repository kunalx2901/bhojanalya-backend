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

// --------------------------------------------------
// Create menu upload entry
// --------------------------------------------------
func (r *PostgresRepository) CreateUpload(
	restaurantID int,
	objectKey string, // ✅ R2 object key
	filename string,
) (int, error) {

	var id int
	err := r.db.QueryRow(
		context.Background(),
		`
		INSERT INTO menu_uploads (
			restaurant_id,
			image_url,
			original_filename,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, 'MENU_UPLOADED', now(), now())
		RETURNING id
		`,
		restaurantID,
		objectKey,
		filename,
	).Scan(&id)

	return id, err
}

// --------------------------------------------------
// Save parsed menu + cost-for-two JSON
// --------------------------------------------------
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
		    status = 'PARSED',
		    updated_at = now()
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
