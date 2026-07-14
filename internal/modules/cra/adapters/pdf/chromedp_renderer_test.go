package pdf

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestHtmlToPDF_WithChromium(t *testing.T) {
	if os.Getenv("CHROME_PATH") == "" && os.Getenv("CI") == "" {
		chromeCandidates := []string{"/usr/bin/chromium", "/usr/bin/chromium-browser", "/usr/bin/google-chrome"}
		for _, path := range chromeCandidates {
			if _, err := os.Stat(path); err == nil {
				t.Setenv("CHROME_PATH", path)
				break
			}
		}
	}
	if os.Getenv("CHROME_PATH") == "" {
		t.Skip("CHROME_PATH not set and chromium not found")
	}
	if os.Getenv("HOME") == "" {
		t.Setenv("HOME", "/tmp")
	}

	html := `<!DOCTYPE html><html><body><h1>CRA test</h1></body></html>`
	pdfBytes, err := htmlToPDF(context.Background(), html)
	if err != nil {
		t.Fatalf("htmlToPDF: %v", err)
	}
	if len(pdfBytes) < 100 {
		t.Fatalf("expected non-empty PDF, got %d bytes", len(pdfBytes))
	}
	if !strings.HasPrefix(string(pdfBytes), "%PDF") {
		t.Fatalf("expected PDF header, got %q", string(pdfBytes[:min(8, len(pdfBytes))]))
	}
}
