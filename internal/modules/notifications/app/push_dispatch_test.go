package app

import "testing"

func TestPushBodyForNotification(t *testing.T) {
	body := pushBodyForNotification("Sujet", "Ligne 1\n\n--\nSignature Kore")
	if body != "Ligne 1" {
		t.Fatalf("expected body without signature, got %q", body)
	}
	long := stringsRepeat("a", 250)
	got := pushBodyForNotification("S", long)
	if len(got) != pushBodyMaxLen {
		t.Fatalf("expected truncated body len %d, got %d", pushBodyMaxLen, len(got))
	}
}

func stringsRepeat(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}
