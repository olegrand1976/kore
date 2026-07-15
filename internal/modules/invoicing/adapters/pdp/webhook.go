package pdp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// VerifyWebhook valide la signature HMAC-SHA256 d'un webhook PDP.
// Header attendu : X-PDP-Signature (hex brut ou préfixe sha256=).
func VerifyWebhook(payload []byte, signature, secret string) bool {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return true
	}
	signature = strings.TrimSpace(signature)
	if signature == "" {
		return false
	}
	if strings.HasPrefix(strings.ToLower(signature), "sha256=") {
		signature = signature[7:]
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(signature)))
}
