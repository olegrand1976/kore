package http

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

// verifyTaigaWebhookAuth accepts either:
// - X-Taiga-Webhook-Secret (smoke tests / relais manuel)
// - X-TAIGA-WEBHOOK-SIGNATURE (webhooks natifs Taiga, HMAC-SHA1 du corps brut)
func verifyTaigaWebhookAuth(r *http.Request, body []byte, webhookSecret string) bool {
	if webhookSecret == "" {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Taiga-Webhook-Secret")), []byte(webhookSecret)) == 1 {
		return true
	}
	signature := r.Header.Get("X-TAIGA-WEBHOOK-SIGNATURE")
	if signature == "" {
		return false
	}
	mac := hmac.New(sha1.New, []byte(webhookSecret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
