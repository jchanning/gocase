package docparse

import (
	"testing"
)

func TestExtractText_UnsupportedExtension(t *testing.T) {
	_, err := ExtractText("notes.docx")
	if err == nil {
		t.Fatal("expected error for unsupported extension")
	}
}

func TestExtractText_NonexistentFile(t *testing.T) {
	_, err := ExtractText("nonexistent.pdf")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
