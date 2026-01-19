package ocr

import (
	"io"
	"net/http"
	"os"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ProcessOne() error {
	id, url, err := s.repo.FetchNext()
	if err != nil {
		return err
	}

	_ = s.repo.UpdateStatus(id, "OCR_PROCESSING", nil)

	resp, err := http.Get(url)
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return err
	}
	defer resp.Body.Close()

	tmpFile := "/tmp/menu_file"
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	_, _ = io.Copy(out, resp.Body)

	text, err := ExtractText(tmpFile)
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return err
	}

	return s.repo.SaveOCRText(id, text)
}
