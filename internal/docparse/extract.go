package docparse

import (
	"fmt"
	"path/filepath"
	"strings"
)

// MaxExtractedLength is the maximum number of characters to return from text extraction.
// This keeps LLM prompts within typical context-window limits.
const MaxExtractedLength = 48000

// ExtractText reads a PDF or PPTX file and returns its plain-text content.
func ExtractText(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	var text string
	var err error

	switch ext {
	case ".pdf":
		text, err = extractPDF(filePath)
	case ".pptx":
		text, err = extractPPTX(filePath)
	default:
		return "", fmt.Errorf("unsupported file type: %s (only .pdf and .pptx are supported)", ext)
	}

	if err != nil {
		return "", fmt.Errorf("text extraction failed for %s: %w", ext, err)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("no text could be extracted from the document")
	}

	// Truncate to stay within LLM context limits
	if len(text) > MaxExtractedLength {
		text = text[:MaxExtractedLength]
	}

	return text, nil
}
