package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
)

// ExtractAttachmentText returns plain text for indexable attachments.
// Images, office and zip return empty string without error (skip index).
func ExtractAttachmentText(storagePath, fileName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".txt", ".md", ".csv", ".log":
		data, err := os.ReadFile(storagePath)
		if err != nil {
			return "", err
		}
		if !utf8.Valid(data) {
			return "", fmt.Errorf("attachment text is not valid utf-8")
		}
		return string(data), nil
	case ".pdf":
		return extractPDFText(storagePath)
	default:
		return "", nil
	}
}

func extractPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	if _, err := buf.ReadFrom(b); err != nil {
		return "", err
	}
	return buf.String(), nil
}
