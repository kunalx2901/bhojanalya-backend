package ocr

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FetchNext() (int, string, error) {
	var id int
	var imageKey string

	err := r.db.QueryRow(
		context.Background(),
		`
		SELECT id, image_url
		FROM menu_uploads
		WHERE status IN ('MENU_UPLOADED', 'OCR_FAILED', 'PARSING_FAILED')
		  AND retry_count < 3
		ORDER BY created_at
		LIMIT 1
		`,
	).Scan(&id, &imageKey)

	if err != nil {
		return 0, "", err
	}

	return id, imageKey, nil
}


func (r *Repository) UpdateStatus(id int, status string, errMsg *string) error {
	_, err := r.db.Exec(
		context.Background(),
		`
		UPDATE menu_uploads
		SET status = $1,
		    ocr_error = $2,
		    retry_count = retry_count + 
		        CASE 
		            WHEN $1 IN ('OCR_FAILED', 'PARSING_FAILED') THEN 1
		            ELSE 0
		        END,
		    updated_at = NOW()
		WHERE id = $3
		`,
		status,
		errMsg,
		id,
	)
	return err
}


// Updated: Only saves text, doesn't set status to DONE yet
func (r *Repository) SaveOCRText(id int, text string) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE menu_uploads 
		SET raw_text = $1, 
			updated_at = now() 
		WHERE id = $2
	`, text, id)

	return err
}

// NEW: Saves the Llama-3.1-8B generated JSON with your business logic
func (r *Repository) SaveStructuredData(id int, jsonData string) error {
	_, err := r.db.Exec(context.Background(), `
		UPDATE menu_uploads 
		SET structured_data = $1,
			status = 'OCR_DONE',
			updated_at = now() 
		WHERE id = $2
	`, jsonData, id)

	return err
}