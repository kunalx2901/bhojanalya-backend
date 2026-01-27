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

// --------------------------------------------------
// Fetch city + cuisine for a menu upload
// --------------------------------------------------
func (r *PostgresRepository) GetMenuContext(
	ctx context.Context,
	menuUploadID int,
) (city string, cuisine string, err error) {

	err = r.db.QueryRow(ctx, `
		SELECT
			r.city,
			r.cuisine_type
		FROM menu_uploads mu
		JOIN restaurants r ON mu.restaurant_id = r.id
		WHERE mu.id = $1
	`, menuUploadID).Scan(&city, &cuisine)

	return
}

// --------------------------------------------------
// ADMIN APPROVAL — FINAL PHASE
// --------------------------------------------------

// List menus that are parsed but not yet approved
func (r *PostgresRepository) ListPending(
	ctx context.Context,
) ([]MenuUpload, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			restaurant_id,
			original_filename,
			parsed_data
		FROM menu_uploads
		WHERE status = 'PARSED'
		  AND approved_at IS NULL
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []MenuUpload

	for rows.Next() {
		var m MenuUpload
		err := rows.Scan(
			&m.ID,
			&m.RestaurantID,
			&m.Filename,
			&m.ParsedData,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}

	return menus, nil
}

// Approve a parsed menu
func (r *PostgresRepository) Approve(
	ctx context.Context,
	menuID int,
	adminID string,
) error {

	_, err := r.db.Exec(ctx, `
		UPDATE menu_uploads
		SET approved_at = now(),
		    approved_by = $2
		WHERE id = $1
	`, menuID, adminID)

	return err
}

// Reject a parsed menu with reason
func (r *PostgresRepository) Reject(
	ctx context.Context,
	menuID int,
	adminID string,
	reason string,
) error {

	_, err := r.db.Exec(ctx, `
		UPDATE menu_uploads
		SET status = 'REJECTED',
		    approved_by = $2,
		    rejection_reason = $3
		WHERE id = $1
	`, menuID, adminID, reason)

	return err
}
