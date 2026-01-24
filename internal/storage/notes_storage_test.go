package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveNotesFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := "test_uploads"
	originalNotesDir := notesDir

	// Override notesDir for testing
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	tests := []struct {
		name          string
		filename      string
		content       string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid PDF file",
			filename:    "test.pdf",
			content:     "PDF content",
			expectError: false,
		},
		{
			name:        "Valid PPT file",
			filename:    "test.ppt",
			content:     "PPT content",
			expectError: false,
		},
		{
			name:        "Valid PPTX file",
			filename:    "test.pptx",
			content:     "PPTX content",
			expectError: false,
		},
		{
			name:          "Invalid file type - TXT",
			filename:      "test.txt",
			content:       "Text content",
			expectError:   true,
			errorContains: "invalid file type",
		},
		{
			name:          "Invalid file type - DOCX",
			filename:      "test.docx",
			content:       "Doc content",
			expectError:   true,
			errorContains: "invalid file type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.content))

			filename, err := SaveNotesFile(reader, tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if filename == "" {
					t.Error("Expected filename to be returned")
				}

				// Check file was actually created
				filePath := GetNotesFilePath(filename)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("File was not created at %s", filePath)
				}

				// Clean up
				DeleteNotesFile(filename)
			}
		})
	}

	// Restore original notesDir
	_ = originalNotesDir
}

func TestDeleteNotesFile(t *testing.T) {
	// Create a test file
	reader := bytes.NewReader([]byte("test content"))
	filename, err := SaveNotesFile(reader, "test.pdf")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify file exists
	filePath := GetNotesFilePath(filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Test file was not created")
	}

	// Delete the file
	err = DeleteNotesFile(filename)
	if err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}

	// Verify file no longer exists
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("File still exists after deletion")
	}

	// Test deleting non-existent file (should not error)
	err = DeleteNotesFile("nonexistent.pdf")
	if err != nil {
		t.Errorf("Deleting non-existent file should not error: %v", err)
	}

	// Test deleting empty filename (should not error)
	err = DeleteNotesFile("")
	if err != nil {
		t.Errorf("Deleting empty filename should not error: %v", err)
	}
}

func TestGetNotesFilePath(t *testing.T) {
	filename := "test-file.pdf"
	expectedPath := filepath.Join(notesDir, filename)

	actualPath := GetNotesFilePath(filename)

	if actualPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, actualPath)
	}
}

func TestIsValidNotesFileType(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".pdf", true},
		{".ppt", true},
		{".pptx", true},
		{".PDF", false}, // Case sensitive
		{".txt", false},
		{".docx", false},
		{".jpg", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := isValidNotesFileType(tt.ext)
			if result != tt.expected {
				t.Errorf("For extension %s, expected %v, got %v", tt.ext, tt.expected, result)
			}
		})
	}
}
