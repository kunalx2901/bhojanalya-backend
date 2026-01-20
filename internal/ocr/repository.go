package ocr

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles OCR-related database operations
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new OCR repository instance
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FetchPending retrieves and CLAIMS the next menu upload pending OCR
// Returns (0, "", nil) when no jobs are available (NOT an error)
func (r *Repository) FetchPending() (int, string, error) {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback(ctx)

	var id int
	var url string

	err = tx.QueryRow(ctx, `
		SELECT id, image_url
		FROM menu_uploads
		WHERE status = 'MENU_UPLOADED'
		ORDER BY created_at
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`).Scan(&id, &url)

	// No pending jobs is NOT an error
	if err != nil {
		return 0, "", nil
	}

	// Mark as processing immediately (atomic claim)
	_, err = tx.Exec(ctx, `
		UPDATE menu_uploads
		SET status = 'OCR_PROCESSING', updated_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return 0, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, "", err
	}

	return id, url, nil
}

// UpdateStatus updates the OCR processing status
func (r *Repository) UpdateStatus(id int, status string, errMsg *string) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE menu_uploads
		SET status = $1,
		    ocr_error = $2,
		    updated_at = now()
		WHERE id = $3
	`, status, errMsg, id)

	return err
}

// SaveText saves extracted OCR text and marks job as complete
func (r *Repository) SaveText(id int, text string) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE menu_uploads
		SET raw_text = $1,
		    status = 'OCR_DONE',
		    updated_at = now()
		WHERE id = $2
	`, text, id)

	return err
}
