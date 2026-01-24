package ocr

import (
	"context"
	"database/sql"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

//
// ─────────────────────────────────────────────────────────────
//  OCR FETCH (MENU_UPLOADED → OCR_PROCESSING)
// ─────────────────────────────────────────────────────────────
//

func (r *Repository) FetchPending() (int, string, error) {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback(ctx)

	var id int
	var imageURL string

	err = tx.QueryRow(ctx, `
		SELECT id, image_url
		FROM menu_uploads
		WHERE status = 'MENU_UPLOADED'
		ORDER BY created_at
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`).Scan(&id, &imageURL)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", sql.ErrNoRows
		}
		return 0, "", err
	}

	// Mark as processing immediately (atomic claim)
	_, err = tx.Exec(ctx, `
		UPDATE menu_uploads
		SET status = 'OCR_PROCESSING',
		    updated_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return 0, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, "", err
	}

	return id, imageURL, nil
}

//
// ─────────────────────────────────────────────────────────────
//  OCR SAVE
// ─────────────────────────────────────────────────────────────
//

func (r *Repository) SaveOCRText(id int, text string) error {
	_, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET raw_text = $1,
		    status = 'OCR_DONE',
		    error_message = NULL,
		    updated_at = now()
		WHERE id = $2
		`,
		text,
		id,
	)
	return err
}

//
// ─────────────────────────────────────────────────────────────
//  LLM FETCH (OCR_DONE → PARSING_LLM)
// ─────────────────────────────────────────────────────────────
//

func (r *Repository) FetchForLLMParsing() (int, string, error) {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback(ctx)

	var id int
	var rawText string

	err = tx.QueryRow(ctx, `
		SELECT id, raw_text
		FROM menu_uploads
		WHERE status = 'OCR_DONE'
		  AND raw_text IS NOT NULL
		  AND parsed_data IS NULL
		ORDER BY updated_at
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`).Scan(&id, &rawText)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", sql.ErrNoRows
		}
		return 0, "", err
	}

	// Mark as parsing
	_, err = tx.Exec(ctx, `
		UPDATE menu_uploads
		SET status = 'PARSING_LLM',
		    updated_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return 0, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, "", err
	}

	return id, rawText, nil
}

//
// ─────────────────────────────────────────────────────────────
//  STATUS UPDATE (GENERIC)
// ─────────────────────────────────────────────────────────────
//

func (r *Repository) UpdateStatus(id int, status string, errMsg *string) error {
	_, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET status = $1,
		    error_message = $2,
		    updated_at = now()
		WHERE id = $3
		`,
		status,
		errMsg,
		id,
	)
	return err
}

//
// ─────────────────────────────────────────────────────────────
//  SAVE PARSED DATA (FINAL STEP)
// ─────────────────────────────────────────────────────────────
//

func (r *Repository) SaveParsedData(id int, parsed any) error {
	_, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET parsed_data = $1,
		    status = 'PARSED',
		    error_message = NULL,
		    updated_at = now()
		WHERE id = $2
		`,
		parsed,
		id,
	)
	return err
}

//
// ─────────────────────────────────────────────────────────────
//  FAILURE HANDLING
// ─────────────────────────────────────────────────────────────
//

func (r *Repository) MarkFailed(id int, reason string) error {
	_, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET status = 'FAILED',
		    error_message = $1,
		    updated_at = now()
		WHERE id = $2
		`,
		reason,
		id,
	)
	return err
}

//
// ─────────────────────────────────────────────────────────────
//  DEBUG / VISIBILITY
// ─────────────────────────────────────────────────────────────
//

func (r *Repository) LogState(id int) {
	var status string
	_ = r.db.QueryRow(
		context.Background(),
		`SELECT status FROM menu_uploads WHERE id = $1`,
		id,
	).Scan(&status)

	log.Printf("📌 menu_uploads[%d] status = %s", id, status)
}
