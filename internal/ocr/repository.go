package ocr

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FetchNext() (int, string, error) {
	var id int
	var url string

	err := r.db.QueryRow(`
		SELECT id, image_url
		FROM menu_uploads
		WHERE status = 'MENU_UPLOADED'
		ORDER BY created_at
		LIMIT 1
	`).Scan(&id, &url)

	return id, url, err
}

func (r *Repository) UpdateStatus(id int, status string, errMsg *string) error {
	_, err := r.db.Exec(`
		UPDATE menu_uploads
		SET status = $1,
		    ocr_error = $2,
		    updated_at = now()
		WHERE id = $3
	`, status, errMsg, id)

	return err
}

func (r *Repository) SaveOCRText(id int, text string) error {
	_, err := r.db.Exec(`
		UPDATE menu_uploads
		SET raw_text = $1,
		    status = 'OCR_DONE',
		    updated_at = now()
		WHERE id = $2
	`, text, id)

	return err
}
