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

	"bhojanalya/internal/competition"
	"bhojanalya/internal/llm"
	"bhojanalya/internal/menu"
	"bhojanalya/internal/storage"
)

type Service struct {
	repo            *Repository
	r2              *storage.R2Client
	llmClient       *llm.GeminiClient
	menuService     *menu.Service
	competitionSvc  *competition.Service
	pdfPreprocessor *PDFTextPreprocessor
}

func NewService(
	repo *Repository,
	r2 *storage.R2Client,
	llmClient *llm.GeminiClient,
	menuService *menu.Service,
	competitionSvc *competition.Service,
) *Service {
	return &Service{
		repo:            repo,
		r2:              r2,
		llmClient:       llmClient,
		menuService:     menuService,
		competitionSvc:  competitionSvc,
		pdfPreprocessor: NewPDFTextPreprocessor(),
	}
}

// ─────────────────────────────────────────────
// OCR WORKER
// ─────────────────────────────────────────────

func (s *Service) RunOCRWorker() {
	log.Println("[OCR WORKER] Started")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.processOCR(); err != nil {
			log.Println("[OCR WORKER] Error:", err)
		}
	}
}

func (s *Service) processOCR() error {
	id, objectKey, err := s.repo.FetchPending()
	if err != nil || id == 0 {
		return nil
	}
	restaurantID, err := s.repo.GetRestaurantID(id)
	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(id, "OCR_FAILED", &msg)
		return nil
	}


	log.Printf("[OCR][%d] Processing restaurant %d", id,restaurantID)
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
	log.Printf("[OCR][%d] OCR completed", id)

	return nil
}

// ─────────────────────────────────────────────
// LLM WORKER (ATOMIC PARSING)
// ─────────────────────────────────────────────

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
	restaurantID, err := s.repo.GetRestaurantID(id)
	if err != nil {
		s.failParsing(id, 0, err)
		return nil
	}


	log.Printf("[LLM][%d] Parsing restaurant %d", id, restaurantID)
	_ = s.repo.UpdateStatus(id, "PARSING_LLM", nil)

	textToParse := rawText
	if s.pdfPreprocessor.IsLikelyPDFText(rawText) {
		textToParse = s.pdfPreprocessor.CleanPDFText(rawText)
	}

	rawJSON, err := s.llmClient.ParseOCR(ctx, textToParse)
	if err != nil {
		s.failParsing(id, restaurantID, err)
		return nil
	}

	parsedOCR, err := llm.ParseLLMResponse(rawJSON)
	if err != nil {
		s.failParsing(id, restaurantID, err)
		return nil
	}

	parsedMenu := toParsedMenu(parsedOCR)

	cost, err := menu.BuildCostForTwo(parsedMenu)
	if err != nil {
		s.failParsing(id, restaurantID, err)
		return nil
	}

	// 🔒 ATOMIC WRITE — THIS IS THE FIX
	if err := s.menuService.SaveParsedResult(
		ctx,
		restaurantID,
		parsedMenu,
		cost,
	); err != nil {
		s.failParsing(id, restaurantID, err)
		return nil
	}

	_ = s.repo.UpdateStatus(id, "PARSED", nil)

	city, cuisine, err := s.menuService.GetMenuContext(ctx, restaurantID)
	if err == nil {
		_ = s.competitionSvc.RecomputeSnapshot(ctx, city, cuisine)
	}

	log.Printf("[PIPELINE][%d] Menu parsed successfully ✅", id)
	return nil
}

func (s *Service) failParsing(
	menuID int,
	restaurantID int,
	err error,
) {
	msg := err.Error()
	_ = s.repo.UpdateStatus(menuID, "PARSING_FAILED", &msg)
	_ = s.menuService.MarkParsingFailed(
		context.Background(),
		restaurantID,
		msg,
	)
}

// ─────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────

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

	cmd := exec.Command("pdftoppm", "-png", pdfPath, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pdftoppm failed: %s", string(out))
	}

	images, err := filepath.Glob(prefix + "*.png")
	if err != nil || len(images) == 0 {
		return "", fmt.Errorf("no images generated from PDF")
	}
	sort.Strings(images)

	var b strings.Builder
	for _, img := range images {
		txt, err := runTesseract(img)
		if err == nil {
			b.WriteString(txt)
			b.WriteString("\n---PAGE BREAK---\n")
		}
		_ = os.Remove(img)
	}

	if b.Len() == 0 {
		return "", fmt.Errorf("no text extracted from PDF")
	}

	return b.String(), nil
}
