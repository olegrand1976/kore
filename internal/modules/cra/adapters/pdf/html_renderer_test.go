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

func TestHTMLRenderer_DetailColumnsMirrorTheGrid(t *testing.T) {
	r := NewHTMLRenderer()
	html, err := r.Render(context.Background(), domain.Timesheet{
		UserID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Month:  "2026-09",
		Status: domain.StatusBrouillon,
	}, CRABrandData{
		Client:        "ACME",
		Mission:       "Support",
		Collaborateur: "DURLET Pascal",
		Lines: []CRALine{{
			Day:        "01/09/2026",
			Person:     "DURLET Pascal",
			Assignment: "Service utilisateur — Hotline",
			Comment:    "Analyse du lot 3",
			Duration:   "6 h 15",
		}},
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got := string(html)
	for _, want := range []string{
		">Jour<", ">Personne<", ">Affectation<", ">Commentaire<", ">Temps<",
		"DURLET Pascal", "Service utilisateur — Hotline", "Analyse du lot 3", "6 h 15",
		"<th>Collaborateur</th>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in CRA HTML", want)
		}
	}
	if strings.Contains(got, "<th>Tâche</th>") {
		t.Fatal("legacy source column still present in CRA HTML")
	}
}
