package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	notesDir = "uploads/notes"
)

// SaveNotesFile saves a notes file (PDF or PowerPoint) and returns the stored filename
func SaveNotesFile(file io.Reader, originalFilename string) (string, error) {
	// Validate file type
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if !isValidNotesFileType(ext) {
		return "", fmt.Errorf("invalid file type: only PDF and PowerPoint files are allowed")
	}

	// Create notes directory if it doesn't exist
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(notesDir, filename)

	// Create the file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Copy data to file
	if _, err := io.Copy(outFile, file); err != nil {
		os.Remove(filePath) // Clean up on error
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filename, nil
}

// GetNotesFilePath returns the full path to a notes file
func GetNotesFilePath(filename string) string {
	return filepath.Join(notesDir, filename)
}

// DeleteNotesFile removes a notes file from storage
func DeleteNotesFile(filename string) error {
	if filename == "" {
		return nil // Nothing to delete
	}

	filePath := filepath.Join(notesDir, filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete notes file: %w", err)
	}

	return nil
}

// isValidNotesFileType checks if the file extension is PDF or PowerPoint
func isValidNotesFileType(ext string) bool {
	validExtensions := map[string]bool{
		".pdf":  true,
		".ppt":  true,
		".pptx": true,
	}
	return validExtensions[ext]
}
