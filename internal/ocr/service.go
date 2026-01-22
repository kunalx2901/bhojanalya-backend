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
	"bhojanalya/internal/menu"
	"bhojanalya/internal/storage"
)

type Service struct {
	repo        *Repository
	r2          *storage.R2Client
	llmClient   *llm.GeminiClient
	menuService *menu.Service
}

func NewService(repo *Repository, r2 *storage.R2Client, llmClient *llm.GeminiClient, menuService *menu.Service) *Service {
	return &Service{
		repo:        repo,
		r2:          r2,
		llmClient:   llmClient,
		menuService: menuService,
	}
}

// Start runs the OCR AND parsing workers forever
func (s *Service) Start() error {
	// Start both workers in separate goroutines
	go s.runOCRWorker()
	go s.runParsingWorker()
	
	// Block forever
	select {}
}

// runOCRWorker processes new menu uploads (OCR phase)
func (s *Service) runOCRWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		err := s.processOne()
		if err != nil {
			log.Println("OCR worker error:", err)
		}
	}
}

// runParsingWorker processes completed OCR results (Parsing phase)
func (s *Service) runParsingWorker() {
	ticker := time.NewTicker(7 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		err := s.processParsingPhase()
		if err != nil {
			log.Println("Parsing worker error:", err)
		}
	}
}

// ProcessOne processes a single OCR task (public API for cmd/ocr-worker)
func (s *Service) ProcessOne() error {
	return s.processOne()
}

func (s *Service) processOne() error {
	log.Println("OCR worker checking for MENU_UPLOADED rows...")

	id, objectKey, err := s.repo.FetchPending()
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

func (s *Service) processParsingPhase() error {
    log.Println("Parsing worker checking for OCR_DONE rows...")
    
    // Now returns single record or nil
    records, err := s.repo.FetchPendingForParsing()
    if err != nil {
        return fmt.Errorf("fetch pending for parsing: %w", err)
    }
    
    if records == nil || len(records) == 0 {
        return nil // No work to do
    }
    
    // Process the claimed record
    rec := records[0]
    log.Printf("Processing OCR record %d for parsing (text length: %d)", 
        rec.ID, len(rec.RawText))
    
    if err := s.ProcessOCR(rec); err != nil {
        log.Printf("Failed to parse OCR record %d: %v", rec.ID, err)
        return nil // Don't return error to keep worker running
    }
    
    log.Printf("✅ Successfully parsed menu upload %d", rec.ID)
    return nil
}

// ProcessOCR processes a single OCR record through LLM
func (s *Service) ProcessOCR(rec OCRRecord) error {
	ctx := context.Background()

	rawJSON, err := s.llmClient.ParseOCR(ctx, rec.RawText)
	if err != nil {
		_ = s.repo.MarkFailed(rec.ID, fmt.Sprintf("LLM parsing failed: %v", err))
		return err
	}

	parsed, err := llm.ParseLLMResponse(rawJSON)
	if err != nil {
		_ = s.repo.MarkFailed(rec.ID, fmt.Sprintf("LLM response parsing failed: %v", err))
		return err
	}

	cost, err := menu.BuildCostForTwo(parsed)
	if err != nil {
		_ = s.repo.MarkFailed(rec.ID, fmt.Sprintf("Cost calculation failed: %v", err))
		return err
	}

	// Create document that matches what menu.SaveParsedMenu expects
	doc := map[string]interface{}{
		"items":        parsed.Items,
		"tax_percent":  parsed.TaxPercent,
		"cost_for_two": cost,
		"ocr_raw_text": rec.RawText, // Keep raw text for debugging
		"parsed_at":    time.Now().UTC(),
		"version":      "1.0",
	}

	// This will update status to 'PARSED' via menu repository
	if err := s.menuService.SaveParsedMenu(rec.ID, doc); err != nil {
		_ = s.repo.MarkFailed(rec.ID, fmt.Sprintf("Failed to save parsed data: %v", err))
		return err
	}

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

// DebugPipeline is a temporary method to verify the pipeline
func (s *Service) DebugPipeline() {
	log.Println("=== DEBUG OCR PIPELINE ===")
	
	// Check how many OCR pending
	id, url, err := s.repo.FetchPending()
	log.Printf("OCR pending (MENU_UPLOADED): id=%d, url=%s, err=%v", id, url, err)
	
	// Check how many parsing pending
	records, err := s.repo.FetchPendingForParsing()
	if err != nil {
		log.Printf("Error fetching parsing pending: %v", err)
	} else if records == nil || len(records) == 0 {
		log.Printf("Parsing pending (OCR_DONE): 0 records (no work)")
	} else {
		log.Printf("Parsing pending (OCR_DONE): %d records", len(records))
		
		// Check each record
		for _, rec := range records {
			log.Printf("  - ID: %d, RawText length: %d", rec.ID, len(rec.RawText))
			
			// Show preview of text
			if len(rec.RawText) > 0 {
				preview := rec.RawText
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				log.Printf("    Preview: %s", preview)
			}
		}
	}
	
	// Also check database directly for broader view
	ctx := context.Background()
	
	// Count by status
	query := `
		SELECT status, COUNT(*) 
		FROM menu_uploads 
		WHERE status IN ('MENU_UPLOADED', 'OCR_PROCESSING', 'OCR_DONE', 'PARSING', 'PARSED')
		GROUP BY status
		ORDER BY status
	`
	
	rows, err := s.repo.db.Query(ctx, query)
	if err != nil {
		log.Printf("Database query error: %v", err)
	} else {
		defer rows.Close()
		
		log.Println("Database Status Summary:")
		for rows.Next() {
			var status string
			var count int
			rows.Scan(&status, &count)
			log.Printf("  %-15s: %d", status, count)
		}
	}
	
	// Check specific record 51
	var status51, rawText51 string
	var parsedData51 interface{}
	err = s.repo.db.QueryRow(ctx, `
		SELECT status, raw_text, parsed_data
		FROM menu_uploads WHERE id = 51
	`).Scan(&status51, &rawText51, &parsedData51)
	
	if err != nil {
		log.Printf("Record 51 query error: %v", err)
	} else {
		hasParsed := "NO"
		if parsedData51 != nil {
			hasParsed = "YES"
		}
		log.Printf("Record 51: Status=%s, TextLen=%d, Parsed=%s", 
			status51, len(rawText51), hasParsed)
	}
	
	log.Println("=== END DEBUG ===")
}