package uploads

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Store saves an uploaded file under baseDir/tenantID/logo.ext.
func Store(baseDir string, tenantID, societeID uuid.UUID, filename string, r io.Reader) (string, error) {
	_ = societeID
	data, err := ReadAndValidateLogo(r, filename)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(baseDir, tenantID.String()), 0o755); err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	destName := "logo" + ext
	destPath := filepath.Join(baseDir, tenantID.String(), destName)
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("/api/v1/branding/logo/%s", tenantID.String()), nil
}

// Path returns the filesystem path for a tenant logo if it exists.
func Path(baseDir string, tenantID uuid.UUID) (string, bool) {
	dir := filepath.Join(baseDir, tenantID.String())
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "logo") {
			return filepath.Join(dir, e.Name()), true
		}
	}
	return "", false
}
