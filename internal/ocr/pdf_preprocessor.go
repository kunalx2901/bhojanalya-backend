package ocr

import (
	"regexp"
	"strings"
	"log"
)

// PDFTextPreprocessor cleans PDF OCR text before LLM parsing
type PDFTextPreprocessor struct{}

func NewPDFTextPreprocessor() *PDFTextPreprocessor {
	return &PDFTextPreprocessor{}
}

// CleanPDFText processes PDF OCR output to make it LLM-friendly
func (p *PDFTextPreprocessor) CleanPDFText(rawText string) string {
	if rawText == "" {
		return rawText
	}
	
	log.Printf("PDF cleaning: Input length = %d chars", len(rawText))
	
	// Step 1: Remove page break markers (added by processPDFtoOCR)
	text := strings.ReplaceAll(rawText, "---PAGE BREAK---", "\n")
	
	// Step 2: Remove page numbers and headers/footers
	text = p.removePageNumbers(text)
	
	// Step 3: Remove excessive whitespace
	text = p.normalizeWhitespace(text)
	
	// Step 4: Remove common PDF artifacts
	text = p.removePDFArtifacts(text)
	
	// Step 5: Limit text length (safeguard for LLM context)
	text = p.smartTruncate(text)
	
	log.Printf("PDF cleaning: Output length = %d chars (reduced by %d%%)", 
		len(text), 100*(len(rawText)-len(text))/len(rawText))
	
	return text
}

func (p *PDFTextPreprocessor) removePageNumbers(text string) string {
	// Remove common page number patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s*page\s*\d+\s*$`),          // "Page 1"
		regexp.MustCompile(`(?i)^\s*\d+\s*/\s*\d+\s*$`),       // "1/5"
		regexp.MustCompile(`^\s*\d+\s*$`),                     // Standalone numbers
		regexp.MustCompile(`(?i)^\s*confidential\s*$`),        // Headers
		regexp.MustCompile(`(?i)^\s*menu\s*$`),                // Repeated headers
	}
	
	lines := strings.Split(text, "\n")
	var cleanLines []string
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		shouldRemove := false
		
		// Check if line matches any removal pattern
		for _, pattern := range patterns {
			if pattern.MatchString(trimmed) {
				shouldRemove = true
				break
			}
		}
		
		// Also remove very short lines that are likely noise
		if !shouldRemove && len(trimmed) < 3 && trimmed != "" {
			// Keep only if it looks like a price (₹, $, numbers)
			if !regexp.MustCompile(`^[₹$€£]?\d*\.?\d+$`).MatchString(trimmed) {
				shouldRemove = true
			}
		}
		
		if !shouldRemove {
			cleanLines = append(cleanLines, line)
		}
	}
	
	return strings.Join(cleanLines, "\n")
}

func (p *PDFTextPreprocessor) normalizeWhitespace(text string) string {
	// Replace multiple spaces with single space
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	
	// Replace multiple newlines with single newline (max 2)
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	
	// Remove leading/trailing whitespace per line
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	
	return strings.Join(lines, "\n")
}

func (p *PDFTextPreprocessor) removePDFArtifacts(text string) string {
	// Remove OCR artifacts common in PDFs
	artifacts := []string{
		"��", "�", "", // OCR garbage characters
		"http://", "https://", // URLs (menus shouldn't have these)
		"©", "™", "®", // Copyright symbols
	}
	
	for _, artifact := range artifacts {
		text = strings.ReplaceAll(text, artifact, "")
	}
	
	return text
}

func (p *PDFTextPreprocessor) smartTruncate(text string) string {
	// Gemini context limit is ~30K tokens, roughly 20K characters for safety
	const maxLength = 15000 // Conservative limit
	
	if len(text) <= maxLength {
		return text
	}
	
	log.Printf("PDF text too long (%d chars), truncating to %d chars", 
		len(text), maxLength)
	
	// Try to truncate at a logical boundary (paragraph)
	truncated := text[:maxLength]
	
	// Find last paragraph break
	if idx := strings.LastIndex(truncated, "\n\n"); idx > maxLength/2 {
		truncated = truncated[:idx]
		log.Printf("Truncated at paragraph break (now %d chars)", len(truncated))
	}
	
	return truncated
}

// IsLikelyPDFText checks if text appears to be from PDF OCR
func (p *PDFTextPreprocessor) IsLikelyPDFText(text string) bool {
	// Check for PDF-specific artifacts
	indicators := []string{
		"---PAGE BREAK---",
		"Page \\d+",
		"", // Form feed character
	}
	
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	
	// PDF OCR tends to be longer and have more line breaks
	if len(text) > 5000 && strings.Count(text, "\n") > len(text)/100 {
		return true
	}
	
	return false
}