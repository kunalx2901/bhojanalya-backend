package ocr

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

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

func (r *Repository) SaveOCRText(id int, text string) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE menu_uploads
		SET raw_text = $1,
		    status = 'OCR_DONE',
		    updated_at = now()
		WHERE id = $2
	`, text, id)

	return err
}

// FetchPendingForParsing retrieves and CLAIMS the next OCR_DONE record for parsing
// Uses same atomic claim pattern as FetchPending()
func (r *Repository) FetchPendingForParsing() ([]OCRRecord, error) {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Query with FOR UPDATE SKIP LOCKED to claim the record atomically
	query := `
		SELECT id, raw_text 
		FROM menu_uploads 
		WHERE status = 'OCR_DONE' 
		AND raw_text IS NOT NULL 
		AND LENGTH(raw_text) > 10
		AND parsed_data IS NULL
		ORDER BY id DESC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`
	
	var id int
	var rawText string
	
	err = tx.QueryRow(ctx, query).Scan(&id, &rawText)
	if err != nil {
		// No pending jobs is NOT an error
		return nil, nil
	}
	
	// Mark as parsing immediately (atomic claim)
	_, err = tx.Exec(ctx, `
		UPDATE menu_uploads
		SET status = 'PARSING', updated_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	
	log.Printf("✅ Claimed record %d for parsing", id)
	
	// Return as slice for compatibility with existing code
	return []OCRRecord{{ID: id, RawText: rawText}}, nil
}

// MarkFailed marks an OCR record as failed with a reason
func (r *Repository) MarkFailed(id int, reason string) error {
	_, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET status = 'FAILED',
		    ocr_error = $1,
		    updated_at = now()
		WHERE id = $2
		`,
		reason,
		id,
	)
	return err
}