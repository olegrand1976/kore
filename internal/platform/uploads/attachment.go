package uploads

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowedAttachmentExt = map[string]bool{
	".pdf": true, ".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".txt": true, ".csv": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".zip": true, ".log": true, ".md": true,
}

const MaxAttachmentBytes = 10 << 20 // 10 MiB

func ValidateAttachmentFilename(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" || !allowedAttachmentExt[ext] {
		return fmt.Errorf("%w: %s", ErrUnsupportedExt, ext)
	}
	base := filepath.Base(filename)
	if base == "" || strings.Contains(base, "..") {
		return fmt.Errorf("%w: invalid filename", ErrInvalidAttachment)
	}
	return nil
}

func ReadAndValidateAttachment(r io.Reader, filename string) ([]byte, error) {
	if err := ValidateAttachmentFilename(filename); err != nil {
		return nil, err
	}
	limited := io.LimitReader(r, MaxAttachmentBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxAttachmentBytes {
		return nil, ErrAttachmentTooLarge
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: empty file", ErrInvalidAttachment)
	}
	return data, nil
}

func StoreAttachment(baseDir string, tenantID, attachmentID uuid.UUID, filename string, r io.Reader) (string, error) {
	data, err := ReadAndValidateAttachment(r, filename)
	if err != nil {
		return "", err
	}
	safeName := filepath.Base(filename)
	dir := filepath.Join(baseDir, tenantID.String(), "attachments", attachmentID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	destPath := filepath.Join(dir, safeName)
	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return "", err
	}
	return destPath, nil
}

func AttachmentPath(storagePath string) (string, bool) {
	if storagePath == "" {
		return "", false
	}
	info, err := os.Stat(storagePath)
	if err != nil || info.IsDir() {
		return "", false
	}
	return storagePath, true
}
