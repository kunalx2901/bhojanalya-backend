package ocr

import (
	"context"
	"database/sql"

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
//  OCR FETCH (MENU_UPLOADED → OCR)
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

	err = tx.QueryRow(
		ctx,
		`
		SELECT id, image_url
		FROM menu_uploads
		WHERE status = 'MENU_UPLOADED'
		ORDER BY created_at
		LIMIT 1
		FOR UPDATE SKIP LOCKED
		`,
	).Scan(&id, &imageURL)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", sql.ErrNoRows
		}
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

	err = tx.QueryRow(
		ctx,
		`
		SELECT id, raw_text
		FROM menu_uploads
		WHERE status = 'OCR_DONE'
		  AND parsed_data IS NULL
		ORDER BY updated_at
		LIMIT 1
		FOR UPDATE SKIP LOCKED
		`,
	).Scan(&id, &rawText)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", sql.ErrNoRows
		}
		return 0, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, "", err
	}

	return id, rawText, nil
}

//
// ─────────────────────────────────────────────────────────────
//  STATUS + PARSED DATA
// ─────────────────────────────────────────────────────────────
//

func (r *Repository) UpdateStatus(id int, status string, errMsg *string) error {
	_, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET status = $1,
		    ocr_error = $2,
		    updated_at = now()
		WHERE id = $3
		`,
		status,
		errMsg,
		id,
	)
	return err
}

func (r *Repository) SaveParsedData(id int, parsed any) error {
	_, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET parsed_data = $1,
		    updated_at = now()
		WHERE id = $2
		`,
		parsed,
		id,
	)
	return err
}
