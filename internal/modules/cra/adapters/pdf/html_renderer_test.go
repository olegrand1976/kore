package pdf

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
)

func TestHTMLRenderer_UsesPrestationHeading(t *testing.T) {
	r := NewHTMLRenderer()
	html, err := r.Render(context.Background(), domain.Timesheet{
		UserID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Month:  "2026-08",
		Status: domain.StatusBrouillon,
	}, CRABrandData{
		Client:  "ACME",
		Mission: "Support",
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got := string(html)
	if !strings.Contains(got, "Informations de prestation") {
		t.Fatal("expected prestation heading in CRA HTML")
	}
	if strings.Contains(got, "Informations commerciales") {
		t.Fatal("legacy commercial heading still present in CRA HTML")
	}
}
