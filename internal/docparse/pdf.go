package docparse

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// extractPDF extracts plain text from a PDF file.
func extractPDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	numPages := r.NumPage()
	if numPages == 0 {
		return "", fmt.Errorf("pdf has no pages")
	}

	var buf bytes.Buffer
	for i := 1; i <= numPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Skip pages that fail to parse rather than aborting
			continue
		}
		if text != "" {
			buf.WriteString(strings.TrimSpace(text))
			buf.WriteString("\n\n")
		}
	}

	return buf.String(), nil
}
