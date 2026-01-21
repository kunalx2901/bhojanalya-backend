package ocr

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bhojanalya/internal/storage"
)

type Service struct {
	repo   *Repository
	r2     *storage.R2Client
	parser *ParserClient
}

// NewService creates a new OCR service
func NewService(repo *Repository, r2 *storage.R2Client, parser *ParserClient) *Service {
	return &Service{
		repo:   repo,
		r2:     r2,
		parser: parser,
	}
}

// Start runs the OCR worker forever
func (s *Service) Start() error {
	for {
		err := s.processOne()
		if err != nil {
			log.Println("OCR idle or error:", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *Service) processOne() error {
	log.Println("OCR worker checking for MENU_UPLOADED rows...")

	id, objectKey, err := s.repo.FetchNext()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	log.Printf("Picked menu ID %d (R2 key: %s)", id, objectKey)

	// ✅ Correct retry-aware state
	_ = s.repo.UpdateStatus(id, "OCR_IN_PROGRESS", nil)

	ext := strings.ToLower(filepath.Ext(objectKey))
	if ext == "" {
		ext = ".png"
	}

	tempDir := os.TempDir()
	localPath := filepath.Join(tempDir, fmt.Sprintf("menu_%d%s", id, ext))

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return err
	}

	// ⬇️ DOWNLOAD FROM CLOUDFLARE R2
	err = storage.DownloadFromR2(
		context.Background(),
		s.r2.GetClient(),
		s.r2.GetBucket(),
		objectKey,
		localPath,
	)
	if err != nil {
		msg := "R2 download failed: " + err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return err
	}
	defer os.Remove(localPath)

	info, err := os.Stat(localPath)
	if err != nil || info.Size() == 0 {
		msg := "downloaded file missing or empty"
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return fmt.Errorf("%s", msg)
	}

	// --- STEP 1: EXTRACT RAW TEXT WITH TESSERACT ---
	var rawText string
	if ext == ".pdf" {
		rawText, err = s.processPDFtoOCR(id, localPath)
	} else {
		rawText, err = runTesseract(localPath)
	}

	if err != nil {
		msg := "Tesseract failed: " + err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return err
	}

	// Save raw OCR text
	if err := s.repo.SaveOCRText(id, rawText); err != nil {
		return err
	}

	// --- STEP 2: PARSE TEXT INTO JSON WITH LLM ---
	log.Printf("Starting LLM parsing for menu ID %d...", id)

	structuredJSON, err := s.parser.ParseWithLogic(rawText)
	if err != nil {
		msg := "LLM parsing failed: " + err.Error()
		log.Printf("Warning: %s", msg)
		_ = s.repo.UpdateStatus(id, "PARSING_FAILED", &msg)
		return err
	}

	// Save structured JSON
	if err := s.repo.SaveStructuredData(id, structuredJSON); err != nil {
		log.Printf("Failed to save structured JSON: %v", err)
		return err
	}

	// ✅ FINAL SUCCESS STATE
	_ = s.repo.UpdateStatus(id, "OCR_DONE", nil)
	log.Printf("OCR and parsing completed successfully for menu ID %d", id)

	return nil
}

// processPDFtoOCR converts PDF to images and extracts text
func (s *Service) processPDFtoOCR(menuID int, pdfPath string) (string, error) {
	tempDir := os.TempDir()
	imagePath := filepath.Join(tempDir, fmt.Sprintf("menu_%d", menuID))

	cmd := exec.Command("pdftoppm", pdfPath, imagePath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdf conversion failed: %w", err)
	}

	var allText string
	files, err := filepath.Glob(imagePath + "-*.ppm")
	if err != nil {
		return "", err
	}

	for _, imgFile := range files {
		text, err := runTesseract(imgFile)
		if err != nil {
			log.Printf("Error processing %s: %v", imgFile, err)
			continue
		}
		allText += text + "\n"
		_ = os.Remove(imgFile)
	}

	return allText, nil
}

// runTesseract runs tesseract OCR on an image file
func runTesseract(imagePath string) (string, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract failed: %w", err)
	}
	return string(out), nil
}
