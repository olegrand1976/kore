package push

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kore/kore/internal/modules/notifications/ports"
)

func TestLoadCredentialBytesInlineJSON(t *testing.T) {
	raw, err := loadCredentialBytes(`{"client_email":"a@b.iam.gserviceaccount.com","private_key":"x"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "{") {
		t.Fatal("expected json bytes")
	}
}

func TestOAuthFormEncodesJWT(t *testing.T) {
	jwt := "eyJ.test+foo/bar=chars"
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", jwt)
	parsed, err := url.ParseQuery(form.Encode())
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Get("assertion") != jwt {
		t.Fatalf("jwt altered: %q", parsed.Get("assertion"))
	}
}

func TestIsInvalidFCMTokenParsesJSON(t *testing.T) {
	err := errorString(`fcm status 404: {"error":{"status":"NOT_FOUND","message":"unregistered"}}`)
	if !isInvalidFCMToken(err) {
		t.Fatal("expected invalid token from JSON status")
	}
}

type errorString string

func (e errorString) Error() string { return string(e) }

func TestFCMSendAuthErrorPropagates(t *testing.T) {
	client := &FCMClient{
		projectID: "proj",
		sa:        &serviceAccount{ClientEmail: "x@y.z", PrivateKey: "invalid"},
		http:      &http.Client{Timeout: time.Second},
	}
	err := client.Send(context.Background(), []string{"tok"}, ports.PushMessage{Title: "t", Body: "b"})
	if err == nil {
		t.Fatal("expected auth error")
	}
	if !strings.Contains(err.Error(), "fcm auth") {
		t.Fatalf("unexpected error: %v", err)
	}
}
