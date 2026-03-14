package docparse

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// extractPPTX extracts plain text from a PPTX file.
// PPTX files are ZIP archives containing XML slide files.
func extractPPTX(filePath string) (string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("open pptx: %w", err)
	}
	defer r.Close()

	// Collect slide files sorted by name (slide1.xml, slide2.xml, …)
	var slideFiles []*zip.File
	for _, f := range r.File {
		dir := filepath.Dir(f.Name)
		base := filepath.Base(f.Name)
		if dir == "ppt/slides" && strings.HasPrefix(base, "slide") && strings.HasSuffix(base, ".xml") {
			slideFiles = append(slideFiles, f)
		}
	}

	sort.Slice(slideFiles, func(i, j int) bool {
		return slideFiles[i].Name < slideFiles[j].Name
	})

	if len(slideFiles) == 0 {
		return "", fmt.Errorf("no slides found in pptx")
	}

	var buf bytes.Buffer
	for _, sf := range slideFiles {
		text, err := extractSlideText(sf)
		if err != nil {
			continue // skip unparseable slides
		}
		if text != "" {
			buf.WriteString(strings.TrimSpace(text))
			buf.WriteString("\n\n")
		}
	}

	return buf.String(), nil
}

// extractSlideText reads a single slide XML and extracts all <a:t> text runs.
func extractSlideText(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var inText bool

	for {
		tok, err := decoder.Token()
		if err != nil {
			break // EOF or error — we've consumed what we can
		}
		switch t := tok.(type) {
		case xml.StartElement:
			// <a:t> elements contain text runs in OOXML
			if t.Name.Local == "t" {
				inText = true
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
				buf.WriteString(" ")
			}
		case xml.CharData:
			if inText {
				buf.Write(t)
			}
		}
	}

	return buf.String(), nil
}
