package ocr

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bhojanalya/internal/llm"
	"bhojanalya/internal/storage"
)

type Service struct {
	repo      *Repository
	r2        *storage.R2Client
	llmClient *llm.GeminiClient
}

func NewService(
	repo *Repository,
	r2 *storage.R2Client,
	llmClient *llm.GeminiClient,
) *Service {
	return &Service{
		repo:      repo,
		r2:        r2,
		llmClient: llmClient,
	}
}

//
// ─────────────────────────────────────────────────────────────
//  OCR WORKER
// ─────────────────────────────────────────────────────────────
//

func (s *Service) StartOCRWorker() {
	log.Println("[OCR WORKER] Started")

	for {
		if err := s.processOCR(); err != nil {
			log.Println("[OCR WORKER] Error:", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *Service) processOCR() error {
	log.Println("[OCR] Checking for MENU_UPLOADED rows...")

	id, objectKey, err := s.repo.FetchPending()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	log.Printf("[OCR][%d] Picked image: %s", id, objectKey)

	_ = s.repo.UpdateStatus(id, "OCR_PROCESSING", nil)

	ext := strings.ToLower(filepath.Ext(objectKey))
	if ext == "" {
		ext = ".png"
	}

	localPath := filepath.Join(os.TempDir(), fmt.Sprintf("menu_%d%s", id, ext))

	// Download from R2
	if err := storage.DownloadFromR2(
		context.Background(),
		s.r2.GetClient(),
		s.r2.GetBucket(),
		objectKey,
		localPath,
	); err != nil {
		msg := "R2 download failed: " + err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil
	}
	defer os.Remove(localPath)

	var text string
	if ext == ".pdf" {
		t, err := s.processPDFtoOCR(id, localPath)
		if err != nil {
			msg := err.Error()
			_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
			return nil
		}
		text = t
	} else {
		t, err := runTesseract(localPath)
		if err != nil {
			msg := err.Error()
			_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
			return nil
		}
		text = t
	}

	if err := s.repo.SaveOCRText(id, text); err != nil {
		return err
	}

	_ = s.repo.UpdateStatus(id, "OCR_DONE", nil)
	log.Printf("[OCR][%d] OCR completed (%d chars)", id, len(text))

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//  LLM PARSING WORKER (THIS FIXES YOUR ISSUE)
// ─────────────────────────────────────────────────────────────
//

func (s *Service) StartLLMWorker() {
	log.Println("[LLM WORKER] Started")

	for {
		if err := s.processLLM(); err != nil {
			log.Println("[LLM WORKER] Error:", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *Service) processLLM() error {
	ctx := context.Background()

	log.Println("[LLM] Checking for OCR_DONE rows...")

	id, rawText, err := s.repo.FetchForLLMParsing()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	log.Printf("[LLM][%d] Picked row (%d chars)", id, len(rawText))

	_ = s.repo.UpdateStatus(id, "PARSING_LLM", nil)

	// 1️⃣ Call Gemini
	rawJSON, err := s.llmClient.ParseOCR(ctx, rawText)
	if err != nil {
		msg := "LLM parsing failed: " + err.Error()
		log.Printf("[LLM][%d][ERROR] %s", id, msg)
		_ = s.repo.UpdateStatus(id, "PARSING_FAILED", &msg)
		return nil
	}

	// 2️⃣ Validate JSON
	parsed, err := llm.ParseLLMResponse(rawJSON)
	if err != nil {
		msg := "JSON validation failed: " + err.Error()
		log.Printf("[LLM][%d][ERROR] %s", id, msg)
		_ = s.repo.UpdateStatus(id, "PARSING_FAILED", &msg)
		return nil
	}

	// 3️⃣ Save parsed JSON
	if err := s.repo.SaveParsedData(id, parsed); err != nil {
		msg := "DB save failed: " + err.Error()
		log.Printf("[LLM][%d][ERROR] %s", id, msg)
		_ = s.repo.UpdateStatus(id, "PARSING_FAILED", &msg)
		return nil
	}

	_ = s.repo.UpdateStatus(id, "PARSED", nil)
	log.Printf("[LLM][%d] Parsing completed successfully", id)

	return nil
}

//
// ─────────────────────────────────────────────────────────────
//  OCR HELPERS
// ─────────────────────────────────────────────────────────────
//

func runTesseract(path string) (string, error) {
	cmd := exec.Command("tesseract", path, "stdout", "-l", "eng")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Tesseract error:\n%s", string(out))
		return "", fmt.Errorf("tesseract failed")
	}
	return string(out), nil
}

func (s *Service) processPDFtoOCR(id int, pdfPath string) (string, error) {
	tempDir := os.TempDir()
	imagePrefix := filepath.Join(tempDir, fmt.Sprintf("menu_%d_page", id))

	cmd := exec.Command("pdftoppm", pdfPath, imagePrefix, "-png")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pdf convert failed: %s", string(out))
	}

	images, _ := filepath.Glob(imagePrefix + "*.png")
	sort.Strings(images)

	var b strings.Builder
	for _, img := range images {
		txt, err := runTesseract(img)
		if err == nil {
			b.WriteString(txt)
			b.WriteString("\n")
		}
		_ = os.Remove(img)
	}

	if b.Len() == 0 {
		return "", fmt.Errorf("no text extracted from PDF")
	}

	return b.String(), nil
}
