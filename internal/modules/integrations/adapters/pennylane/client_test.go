package pennylane

import "testing"

func TestMonthBounds(t *testing.T) {
	start, end, err := monthBounds("2026-03")
	if err != nil {
		t.Fatal(err)
	}
	if start != "2026-03-01" || end != "2026-03-31" {
		t.Fatalf("unexpected bounds: %s..%s", start, end)
	}
}
