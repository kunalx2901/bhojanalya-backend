package ocr

import (
	"context"
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
	repo            *Repository
	r2              *storage.R2Client
	llmClient       *llm.GeminiClient
	menuService     *menu.Service
	pdfPreprocessor *PDFTextPreprocessor
}

func NewService(
	repo *Repository,
	r2 *storage.R2Client,
	llmClient *llm.GeminiClient,
	menuService *menu.Service,
) *Service {
	return &Service{
		repo:            repo,
		r2:              r2,
		llmClient:       llmClient,
		menuService:     menuService,
		pdfPreprocessor: NewPDFTextPreprocessor(),
	}
}

//
// ─────────────────────────────────────────────
// OCR WORKER
// ─────────────────────────────────────────────
//

func (s *Service) RunOCRWorker() {
	log.Println("[OCR WORKER] Started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.processOCR(); err != nil {
			log.Println("[OCR WORKER] Error:", err)
		}
	}
}

func (s *Service) processOCR() error {
	log.Println("[OCR] Checking MENU_UPLOADED rows")

	id, objectKey, err := s.repo.FetchPending()
	if err != nil || id == 0 {
		return nil
	}

	log.Printf("[OCR][%d] Picked %s", id, objectKey)
	_ = s.repo.UpdateStatus(id, "OCR_PROCESSING", nil)

	ext := strings.ToLower(filepath.Ext(objectKey))
	if ext == "" {
		ext = ".png"
	}

	localPath := filepath.Join(os.TempDir(), fmt.Sprintf("menu_%d%s", id, ext))

	if err := storage.DownloadFromR2(
		context.Background(),
		s.r2.GetClient(),
		s.r2.GetBucket(),
		objectKey,
		localPath,
	); err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil
	}
	defer os.Remove(localPath)

	var text string
	if ext == ".pdf" {
		text, err = s.processPDFtoOCR(id, localPath)
	} else {
		text, err = runTesseract(localPath)
	}

	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil
	}

	if err := s.repo.SaveOCRText(id, text); err != nil {
		return err
	}

	_ = s.repo.UpdateStatus(id, "OCR_DONE", nil)
	log.Printf("[OCR][%d] Done (%d chars)", id, len(text))
	
	if ext == ".pdf" {
		log.Printf("[OCR][%d] PDF text saved, page count: %d", 
			id, strings.Count(text, "\n\n")+1)
	}
	
	return nil
}

//
// ─────────────────────────────────────────────
// LLM + COST-FOR-TWO WORKER
// ─────────────────────────────────────────────
//

func (s *Service) RunLLMWorker() {
	log.Println("[LLM WORKER] Started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.processLLM(); err != nil {
			log.Println("[LLM WORKER] Error:", err)
		}
	}
}

func (s *Service) processLLM() error {
	ctx := context.Background()

	id, rawText, err := s.repo.FetchForLLMParsing()
	if err != nil || id == 0 {
		return nil
	}

	log.Printf("[LLM][%d] Parsing (%d chars)", id, len(rawText))
	_ = s.repo.UpdateStatus(id, "PARSING_LLM", nil)

	textToParse := rawText
	if s.pdfPreprocessor.IsLikelyPDFText(rawText) {
		log.Printf("[LLM][%d] PDF detected, cleaning text...", id)
		textToParse = s.pdfPreprocessor.CleanPDFText(rawText)
		log.Printf("[LLM][%d] PDF cleaning: %d → %d chars", 
			id, len(rawText), len(textToParse))
	}

	rawJSON, err := s.llmClient.ParseOCR(ctx, textToParse)
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "PARSING_FAILED", &msg)
		return nil
	}

	parsedOCR, err := llm.ParseLLMResponse(rawJSON)
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "PARSING_FAILED", &msg)
		return nil
	}

	parsedMenu := toParsedMenu(parsedOCR)

	cost, err := menu.BuildCostForTwo(parsedMenu)
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "PARSING_FAILED", &msg)
		return nil
	}

	if err := s.menuService.SaveParsedResult(id, parsedMenu, cost); err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "PARSING_FAILED", &msg)
		return nil
	}

	_ = s.repo.UpdateStatus(id, "PARSED", nil)
	log.Printf("[LLM][%d] Parsed + cost-for-two saved", id)
	
	// ✅ FIXED: Check if Calculation is not zero value
	if cost.Calculation.Total > 0 {
		log.Printf("[LLM][%d] Success: %d items, tax: %.1f%%, total: %.2f",
			id, len(parsedMenu.Items), parsedMenu.TaxPercent, cost.Calculation.Total)
	} else {
		log.Printf("[LLM][%d] Success: %d items, tax: %.1f%%",
			id, len(parsedMenu.Items), parsedMenu.TaxPercent)
	}
	
	return nil
}

//
// ─────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────
//

func toParsedMenu(ocr *llm.ParsedOCRResult) *menu.ParsedMenu {
	items := make([]menu.Item, 0, len(ocr.Items))

	for _, it := range ocr.Items {
		items = append(items, menu.Item{
			Name:     it.Name,
			Category: it.Category,
			Price:    it.Price,
		})
	}

	return &menu.ParsedMenu{
		Items:      items,
		TaxPercent: ocr.TaxPercent,
	}
}

func runTesseract(path string) (string, error) {
	cmd := exec.Command("tesseract", path, "stdout", "-l", "eng")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tesseract failed: %s", string(out))
	}
	return string(out), nil
}

func (s *Service) processPDFtoOCR(id int, pdfPath string) (string, error) {
	prefix := filepath.Join(os.TempDir(), fmt.Sprintf("menu_%d_page", id))

	cmd := exec.Command("pdftoppm", pdfPath, prefix, "-png")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pdf convert failed: %s", string(out))
	}

	images, _ := filepath.Glob(prefix + "*.png")
	sort.Strings(images)

	var b strings.Builder
	for _, img := range images {
		txt, err := runTesseract(img)
		if err == nil {
			b.WriteString(txt + "\n")
		}
		_ = os.Remove(img)
	}

	if b.Len() == 0 {
		return "", fmt.Errorf("no text extracted from PDF")
	}
	
	result := b.String()
	if len(images) > 1 {
		result = strings.ReplaceAll(result, "\n\n", "\n---PAGE BREAK---\n")
	}
	
	return result, nil
}

func (s *Service) DebugPDFText(text string) string {
	if s.pdfPreprocessor.IsLikelyPDFText(text) {
		log.Println("[DEBUG] PDF text detected")
		cleaned := s.pdfPreprocessor.CleanPDFText(text)
		log.Printf("[DEBUG] Length: %d → %d chars", len(text), len(cleaned))
		
		preview := cleaned
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		log.Printf("[DEBUG] Preview: %s", preview)
		
		return cleaned
	}
	log.Println("[DEBUG] Not PDF text, returning original")
	return text
}

