package menu

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// --------------------------------------------------
// GET MENU STATUS
// --------------------------------------------------

type MenuStatus struct {
	Status string
	Reason *string
}

func (r *PostgresRepository) GetMenuStatus(
	ctx context.Context,
	restaurantID int,
) (*MenuStatus, error) {

	var status string
	var reason *string

	err := r.db.QueryRow(ctx, `
		SELECT status, rejection_reason
		FROM menu_uploads
		WHERE restaurant_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, restaurantID).Scan(&status, &reason)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("no menu uploaded")
		}
		return nil, err
	}

	return &MenuStatus{
		Status: status,
		Reason: reason,
	}, nil
}

// --------------------------------------------------
// UPSERT MENU UPLOAD (ONE MENU PER RESTAURANT)
// --------------------------------------------------
func (r *PostgresRepository) UpsertUpload(
	ctx context.Context,
	restaurantID int,
	objectKey string,
	filename string,
) (int, string, error) {

	var (
		menuID int
		status string
	)

	// Check existing menu (if any)
	err := r.db.QueryRow(ctx, `
		SELECT id, status
		FROM menu_uploads
		WHERE restaurant_id = $1
	`, restaurantID).Scan(&menuID, &status)

	if err == nil {
		// Menu already exists
		if status == "PARSED" {
			return menuID, status, errors.New("menu already parsed and locked")
		}

		// Replace existing (retry allowed)
		_, err := r.db.Exec(ctx, `
			UPDATE menu_uploads
			SET image_url = $1,
			    original_filename = $2,
			    status = 'MENU_UPLOADED',
			    parsed_data = NULL,
			    rejection_reason = NULL,
			    updated_at = now()
			WHERE restaurant_id = $3
		`, objectKey, filename, restaurantID)

		return menuID, "MENU_UPLOADED", err
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, "", err
	}

	// No menu exists → create once
	err = r.db.QueryRow(ctx, `
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
	`, restaurantID, objectKey, filename).Scan(&menuID)

	return menuID, "MENU_UPLOADED", err
}

// --------------------------------------------------
// MARK PARSED (ATOMIC, SAFE)
// --------------------------------------------------
func (r *PostgresRepository) MarkParsed(
	ctx context.Context,
	restaurantID int,
	doc map[string]interface{},
) error {

	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, `
		UPDATE menu_uploads
		SET parsed_data = $1,
		    status = 'PARSED',
		    updated_at = now()
		WHERE restaurant_id = $2
	`, data, restaurantID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("no menu row updated")
	}

	return tx.Commit(ctx)
}

// --------------------------------------------------
// MARK FAILED (NO PARSED DATA WRITTEN)
// --------------------------------------------------
func (r *PostgresRepository) MarkFailed(
	ctx context.Context,
	restaurantID int,
	reason string,
) error {

	_, err := r.db.Exec(ctx, `
		UPDATE menu_uploads
		SET status = 'FAILED',
		    rejection_reason = $1,
		    updated_at = now()
		WHERE restaurant_id = $2
	`, reason, restaurantID)

	return err
}

// --------------------------------------------------
// RETRY FAILED MENU (SAFE RESET)
// --------------------------------------------------
func (r *PostgresRepository) RetryFailedMenu(
	ctx context.Context,
	restaurantID int,
) error {

	cmd, err := r.db.Exec(ctx, `
		UPDATE menu_uploads
		SET status = 'MENU_UPLOADED',
		    parsed_data = NULL,
		    rejection_reason = NULL,
		    updated_at = now()
		WHERE restaurant_id = $1
		  AND status = 'FAILED'
	`, restaurantID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("menu not in FAILED state or not found")
	}

	return nil
}

// --------------------------------------------------
// MENU CONTEXT (FOR COMPETITION SNAPSHOT)
// --------------------------------------------------
func (r *PostgresRepository) GetMenuContext(
	ctx context.Context,
	restaurantID int,
) (city string, cuisine string, err error) {

	err = r.db.QueryRow(ctx, `
		SELECT
			r.city,
			r.cuisine_type
		FROM restaurants r
		WHERE r.id = $1
	`, restaurantID).Scan(&city, &cuisine)

	return
}

// --------------------------------------------------
// ADMIN APPROVAL — FINAL PHASE
// --------------------------------------------------


// List menus pending approval
func (r *PostgresRepository) ListPending(
	ctx context.Context,
) ([]MenuUpload, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			mu.id,
			mu.restaurant_id,
			r.name,
			r.city,
			r.cuisine_type,
			r.opens_at,
			r.closes_at,
			mu.original_filename,
			mu.parsed_data
		FROM menu_uploads mu
		JOIN restaurants r
		  ON r.id = mu.restaurant_id
		WHERE mu.status = 'PARSED'
		  AND mu.approved_at IS NULL
		ORDER BY mu.updated_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []MenuUpload

	for rows.Next() {
		var m MenuUpload
		if err := rows.Scan(
			&m.ID,
			&m.RestaurantID,
			&m.RestaurantName,
			&m.City,
			&m.CuisineType,
			&m.OpensAt,
			&m.ClosesAt,
			&m.Filename,
			&m.ParsedData,
		); err != nil {
			return nil, err
		}

		menus = append(menus, m)
	}

	return menus, nil
}




// Approve menu (ADMIN)
func (r *PostgresRepository) Approve(
	ctx context.Context,
	menuID int,
	adminID string,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1️⃣ Approve menu
	cmd, err := tx.Exec(ctx, `
		UPDATE menu_uploads
		SET approved_at = now(),
		    approved_by = $2
		WHERE id = $1
		  AND status = 'PARSED'
	`, menuID, adminID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("menu not found or not parsed")
	}

	// 2️⃣ Approve restaurant
	_, err = tx.Exec(ctx, `
		UPDATE restaurants
		SET status = 'approved'
		WHERE id = (
			SELECT restaurant_id
			FROM menu_uploads
			WHERE id = $1
		)
	`, menuID)
	if err != nil {
		return err
	}

	// 3️⃣ Approve all pending deals
	_, err = tx.Exec(ctx, `
		UPDATE deals
		SET status = 'APPROVED',
		    approved_at = now(),
		    approved_by = $2
		WHERE restaurant_id = (
			SELECT restaurant_id
			FROM menu_uploads
			WHERE id = $1
		)
		AND status = 'PENDING_APPROVAL'
	`, menuID, adminID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Reject menu (ADMIN)
func (r *PostgresRepository) Reject(
	ctx context.Context,
	restaurantID int,
	adminID string,
	reason string,
) error {

	_, err := r.db.Exec(ctx, `
		UPDATE menu_uploads
		SET status = 'REJECTED',
		    approved_by = $2,
		    rejection_reason = $3,
		    updated_at = now()
		WHERE restaurant_id = $1
	`, restaurantID, adminID, reason)

	return err
}

func (r *PostgresRepository) ApproveByRestaurant(
	ctx context.Context,
	restaurantID int,
	adminID string,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1️⃣ Approve latest PARSED menu
	cmd, err := tx.Exec(ctx, `
		UPDATE menu_uploads
		SET approved_at = now(),
		    approved_by = $2
		WHERE restaurant_id = $1
		  AND status = 'PARSED'
		  AND approved_at IS NULL
	`, restaurantID, adminID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("no parsed menu found for restaurant")
	}

	// 2️⃣ Approve restaurant
	_, err = tx.Exec(ctx, `
		UPDATE restaurants
		SET status = 'approved'
		WHERE id = $1
	`, restaurantID)
	if err != nil {
		return err
	}

	// 3️⃣ Approve all pending deals
	_, err = tx.Exec(ctx, `
		UPDATE deals
		SET status = 'APPROVED',
		    approved_at = now(),
		    approved_by = $2
		WHERE restaurant_id = $1
		  AND status = 'PENDING_APPROVAL'
	`, restaurantID, adminID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
