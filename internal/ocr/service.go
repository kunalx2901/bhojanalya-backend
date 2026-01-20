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

	tmpFile, err := os.CreateTemp("", "menu-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	_, _ = io.Copy(tmpFile, resp.Body)

	text, err := ExtractText(tmpFile.Name())
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return err
	}

	return s.repo.SaveOCRText(id, text)
}
