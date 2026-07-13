package uploads

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

var allowedLogoExt = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
}

const MaxLogoBytes = 512 << 10

func ValidateLogoFilename(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := allowedLogoExt[ext]; !ok {
		return fmt.Errorf("%w: %s", ErrUnsupportedExt, ext)
	}
	return nil
}

func ContentTypeForExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ct, ok := allowedLogoExt[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// ValidateLogoContent checks magic bytes and basic SVG safety on the first chunk.
func ValidateLogoContent(filename string, head []byte) error {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		if len(head) < 8 || !bytes.HasPrefix(head, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
			return fmt.Errorf("%w: invalid png content", ErrInvalidLogo)
		}
	case ".jpg", ".jpeg":
		if len(head) < 3 || head[0] != 0xff || head[1] != 0xd8 || head[2] != 0xff {
			return fmt.Errorf("%w: invalid jpeg content", ErrInvalidLogo)
		}
	case ".webp":
		if len(head) < 12 || string(head[0:4]) != "RIFF" || string(head[8:12]) != "WEBP" {
			return fmt.Errorf("%w: invalid webp content", ErrInvalidLogo)
		}
	case ".svg":
		s := strings.ToLower(strings.TrimSpace(string(head)))
		if !strings.HasPrefix(s, "<svg") && !strings.HasPrefix(s, "<?xml") {
			return fmt.Errorf("%w: invalid svg content", ErrInvalidLogo)
		}
		if strings.Contains(s, "<script") {
			return fmt.Errorf("%w: svg scripts are not allowed", ErrInvalidLogo)
		}
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedExt, ext)
	}
	return nil
}

func ReadAndValidateLogo(r io.Reader, filename string) ([]byte, error) {
	if err := ValidateLogoFilename(filename); err != nil {
		return nil, err
	}
	limited := io.LimitReader(r, MaxLogoBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxLogoBytes {
		return nil, ErrLogoTooLarge
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: empty file", ErrInvalidLogo)
	}
	head := data
	if len(head) > 512 {
		head = head[:512]
	}
	if err := ValidateLogoContent(filename, head); err != nil {
		return nil, err
	}
	if strings.ToLower(filepath.Ext(filename)) == ".svg" {
		lower := strings.ToLower(string(data))
		if strings.Contains(lower, "<script") {
			return nil, fmt.Errorf("%w: svg scripts are not allowed", ErrInvalidLogo)
		}
	}
	return data, nil
}
