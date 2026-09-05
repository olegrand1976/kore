package pdf

import (
	"context"
	"html/template"
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

func TestHTMLRenderer_LetterheadAndProductFooter(t *testing.T) {
	// A data: URI is exactly what html/template's URL filter neutralises into
	// #ZgotmplZ unless the field is typed template.URL — the logo silently
	// disappearing from every PDF is the regression this pins down.
	logo := template.URL("data:image/png;base64,iVBORw0KGgo=")
	html, err := NewHTMLRenderer().Render(context.Background(), domain.Timesheet{
		UserID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Month:  "2026-09",
		Status: domain.StatusBrouillon,
	}, CRABrandData{
		CompanyName:    "LL-IT Software and Computer",
		CompanyAddr:    "Rue de la Résistance 92 / A, 7131 Waudrez, Belgique",
		CompanyLogo:    logo,
		ShowKoreFooter: true,
		ProductSite:    "kore.ll-it-sc.be",
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got := string(html)
	if strings.Contains(got, "ZgotmplZ") {
		t.Fatal("company logo URL was neutralised by html/template")
	}
	for _, want := range []string{
		string(logo),
		"LL-IT Software and Computer",
		"Rue de la Résistance 92 / A, 7131 Waudrez, Belgique",
		`aria-label="Kore"`,
		"Kore — Le cockpit de votre activité IT",
		"kore.ll-it-sc.be",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in CRA HTML", want)
		}
	}
}

func TestHTMLRenderer_WhiteLabelDropsProductFooter(t *testing.T) {
	html, err := NewHTMLRenderer().Render(context.Background(), domain.Timesheet{
		UserID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Month:  "2026-09",
		Status: domain.StatusBrouillon,
	}, CRABrandData{CompanyName: "ACME", ShowKoreFooter: false, ProductSite: "kore.ll-it-sc.be"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got := string(html)
	if strings.Contains(got, "Le cockpit de votre activité IT") || strings.Contains(got, `aria-label="Kore"`) {
		t.Fatal("product footer rendered despite ShowKoreFooter=false")
	}
	if !strings.Contains(got, "Généré le") {
		t.Fatal("generation stamp should survive white-labelling")
	}
}
