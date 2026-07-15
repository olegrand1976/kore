package pdp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyWebhook(t *testing.T) {
	payload := []byte(`{"status":"accepted"}`)
	secret := "test-secret"
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	if !VerifyWebhook(payload, sig, secret) {
		t.Fatal("expected valid signature")
	}
	if !VerifyWebhook(payload, "sha256="+sig, secret) {
		t.Fatal("expected valid sha256= prefixed signature")
	}
	if VerifyWebhook(payload, "bad", secret) {
		t.Fatal("expected invalid signature")
	}
	if !VerifyWebhook(payload, "", "") {
		t.Fatal("empty secret should skip verification in dev")
	}
}
