package ocr

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ProcessOne picks ONE pending OCR job and processes it safely
func (s *Service) ProcessOne() error {
	id, url, err := s.repo.FetchPending()
	if err != nil || id == 0 {
		// No pending jobs is NOT an error
		return nil
	}

	_ = s.repo.UpdateStatus(id, "OCR_PROCESSING", nil)

	resp, err := http.Get(url)
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil // 👈 do NOT block worker
	}
	defer resp.Body.Close()

	log.Printf("OCR_FETCHED id=%d content-type=%s", id, resp.Header.Get("Content-Type"))

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil
	}

	// 🔴 Detect PDF safely
	if bytes.HasPrefix(bodyBytes, []byte("%PDF")) {
		msg := "PDF files not supported yet"
		log.Printf("OCR_SKIPPED (PDF) id=%d url=%s", id, url)
		_ = s.repo.UpdateStatus(id, "OCR_SKIPPED", &msg)
		return nil
	}

	// 🟢 Write image to temp file
	tmpFile, err := os.CreateTemp("", "menu-*.jpg")
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil
	}
	defer os.Remove(tmpFile.Name())

	written, err := io.Copy(tmpFile, bytes.NewReader(bodyBytes))
	if err != nil || written == 0 {
		msg := "failed to write temp image"
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil
	}

	_ = tmpFile.Close()

	log.Printf("OCR_PROCESSING id=%d file=%s bytes=%d", id, tmpFile.Name(), written)

	text, err := ExtractText(tmpFile.Name())
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil
	}

	log.Printf("OCR_DONE id=%d text_length=%d", id, len(text))

	return s.repo.SaveText(id, text)
}
