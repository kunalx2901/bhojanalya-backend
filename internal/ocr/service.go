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

	"bhojanalya/internal/storage"
)

type Service struct {
	repo *Repository
	r2   *storage.R2Client
}


func NewService(repo *Repository, r2 *storage.R2Client) *Service {
	return &Service{
		repo: repo,
		r2:   r2,
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

	_ = s.repo.UpdateStatus(id, "OCR_PROCESSING", nil)

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
		return nil // 👈 do NOT block worker
	}
	defer os.Remove(localPath) // Clean up temp file

	info, err := os.Stat(localPath)
	if err != nil || info.Size() == 0 {
		msg := "downloaded file missing or empty"
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return fmt.Errorf("%s", msg)
	}

	log.Printf("OCR input file ready: %s (%d bytes)", localPath, info.Size())

	// Process based on file type
	var text string
	if ext == ".pdf" {
		// Convert PDF to images and OCR
		var pdfErr error
		text, pdfErr = s.processPDFtoOCR(id, localPath)
		if pdfErr != nil {
			msg := pdfErr.Error()
			_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
			return pdfErr
		}
	} else {
		// Process image directly with Tesseract
		var tesseractErr error
		text, tesseractErr = runTesseract(localPath)
		if tesseractErr != nil {
			msg := tesseractErr.Error()
			_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
			return tesseractErr
		}
	}

	if err := s.repo.SaveOCRText(id, text); err != nil {
		return err
	}

	// Clear any previous OCR error
	_ = s.repo.UpdateStatus(id, "OCR_DONE", nil)

	log.Printf("OCR completed successfully for menu ID %d", id)
	return nil
}

func runTesseract(path string) (string, error) {
	cmd := exec.Command("tesseract", path, "stdout", "-l", "eng")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Tesseract error output:\n%s", string(out))
		return "", fmt.Errorf("tesseract failed")
	}

	return string(out), nil
}

// processPDFtoOCR converts PDF to images and runs OCR on each page
func (s *Service) processPDFtoOCR(id int, pdfPath string) (string, error) {
	tempDir := os.TempDir()
	imagePrefix := filepath.Join(tempDir, fmt.Sprintf("menu_%d_page", id))

	// Convert PDF to PNG images (one per page)
	cmd := exec.Command("pdftoppm", pdfPath, imagePrefix, "-png")
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("PDF conversion failed: %s", string(out))
		return "", fmt.Errorf("failed to convert PDF to images: %w", err)
	}

	// Find all generated image files
	pattern := imagePrefix + "*.png"
	images, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to list generated images: %w", err)
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no images generated from PDF")
	}

	// Sort images by filename (ensures correct page order)
	sort.Strings(images)

	// OCR each generated image and combine results
	var fullText strings.Builder
	for _, imgPath := range images {
		log.Printf("OCR processing PDF page: %s", filepath.Base(imgPath))

		pageText, err := runTesseract(imgPath)
		if err != nil {
			log.Printf("OCR failed on page %s: %v", filepath.Base(imgPath), err)
			// Continue with next page instead of failing entirely
			continue
		}

		fullText.WriteString(pageText)
		fullText.WriteString("\n---PAGE BREAK---\n")

		// Clean up individual page image
		_ = os.Remove(imgPath)
	}

	if fullText.Len() == 0 {
		return "", fmt.Errorf("no text extracted from PDF")
	}

	log.Printf("PDF OCR completed: extracted %d bytes from %d pages", fullText.Len(), len(images))
	return fullText.String(), nil
}

