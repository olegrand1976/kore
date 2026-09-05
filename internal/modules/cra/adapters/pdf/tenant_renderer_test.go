package pdf

import (
	"context"
	"encoding/base64"
	"errors"
	"html/template"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

var testTenant = kernel.NewTenantID(uuid.MustParse("11111111-1111-1111-1111-111111111111"))

type fakeWorkRefReader struct {
	subjects map[uuid.UUID]string
	calls    int
	err      error
}

func (f *fakeWorkRefReader) WorkRefLabel(_ context.Context, _ kernel.TenantID, id uuid.UUID) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.subjects[id], nil
}

type fakeUserResolver struct {
	identity ports.UserIdentity
	err      error
}

func (f fakeUserResolver) ResolveUserIdentity(_ context.Context, _ kernel.TenantID, _ uuid.UUID) (ports.UserIdentity, error) {
	return f.identity, f.err
}

// fakeOrg overrides only the branding call the renderer makes; the embedded
// interface supplies the rest of the method set and panics if anything else is
// reached, which is exactly the signal we want.
type fakeOrg struct {
	orgports.OrganizationService
	logo        []byte
	contentType string
	err         error
}

func (f fakeOrg) GetTenantLogo(_ context.Context, _ kernel.TenantID) ([]byte, string, error) {
	return f.logo, f.contentType, f.err
}

func TestProductSiteHost(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"public https origin", "https://kore.ll-it-sc.be", "kore.ll-it-sc.be"},
		{"strips path and port", "https://kore.ll-it-sc.be:8443/app", "kore.ll-it-sc.be"},
		{"dev origin dropped", "http://localhost:3001", ""},
		{"bare host dropped", "http://kore", ""},
		{"ip dropped", "http://10.0.0.4:3001", ""},
		{"empty", "", ""},
		{"not a url", "kore.ll-it-sc.be", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := productSiteHost(tc.raw); got != tc.want {
				t.Fatalf("productSiteHost(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestImageMIME(t *testing.T) {
	cases := []struct {
		raw    string
		want   string
		wantOK bool
	}{
		{"image/png", "image/png", true},
		{"image/svg+xml; charset=utf-8", "image/svg+xml", true},
		{" IMAGE/JPEG ", "image/jpeg", true},
		{"application/octet-stream", "", false},
		{"text/html", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			got, ok := imageMIME(tc.raw)
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("imageMIME(%q) = (%q, %v), want (%q, %v)", tc.raw, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestCompanyLogo(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', 0x0d}
	wantDataURI := template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(png))

	cases := []struct {
		name     string
		org      orgports.OrganizationService
		declared string
		want     template.URL
	}{
		{"stored logo wins", fakeOrg{logo: png, contentType: "image/png"}, "https://cdn.example/logo.png", wantDataURI},
		{"absolute url fallback", fakeOrg{err: errors.New("no logo")}, "https://cdn.example/logo.png", "https://cdn.example/logo.png"},
		{"api path is not fetchable", fakeOrg{err: errors.New("no logo")}, "/api/v1/org/branding/logo/abc", ""},
		{"unsupported content type", fakeOrg{logo: png, contentType: "application/pdf"}, "", ""},
		{"empty content", fakeOrg{contentType: "image/png"}, "", ""},
		{"no org service", nil, "https://cdn.example/logo.png", "https://cdn.example/logo.png"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &TenantRenderer{org: tc.org}
			if got := r.companyLogo(context.Background(), testTenant, tc.declared); got != tc.want {
				t.Fatalf("companyLogo = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTrimJoin(t *testing.T) {
	cases := []struct {
		name  string
		parts []string
		want  string
	}{
		{"no part", nil, ""},
		{"only empty parts", []string{"", ""}, ""},
		{"single part", []string{"ACME"}, "ACME"},
		{"drops empty parts in between", []string{"ACME", "", "Bruxelles"}, "ACME · Bruxelles"},
		{"all parts", []string{"ACME", "Mission", "Bruxelles"}, "ACME · Mission · Bruxelles"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := trimJoin(tc.parts...); got != tc.want {
				t.Fatalf("trimJoin(%q) = %q, want %q", tc.parts, got, tc.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		name    string
		minutes int
		want    string
	}{
		{"whole hours", 480, "8 h"},
		{"half hour", 450, "7 h 30"},
		{"quarter hour", 375, "6 h 15"},
		{"three quarters below an hour", 45, "0 h 45"},
		{"pads single digit minutes", 365, "6 h 05"},
		{"zero", 0, "0 h"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatDuration(tc.minutes); got != tc.want {
				t.Fatalf("formatDuration(%d) = %q, want %q", tc.minutes, got, tc.want)
			}
		})
	}
}

func TestFormatPerson(t *testing.T) {
	cases := []struct {
		name     string
		identity ports.UserIdentity
		want     string
	}{
		{"nom and prénom", ports.UserIdentity{Nom: "DURLET", Prenom: "Pascal", Login: "COL_pascal"}, "DURLET Pascal"},
		{"nom only", ports.UserIdentity{Nom: "DURLET", Login: "COL_pascal"}, "DURLET"},
		{"falls back to login", ports.UserIdentity{Login: "COL_pascal"}, "COL_pascal"},
		{"empty identity", ports.UserIdentity{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatPerson(tc.identity); got != tc.want {
				t.Fatalf("formatPerson(%+v) = %q, want %q", tc.identity, got, tc.want)
			}
		})
	}
}

func TestSourceLabel(t *testing.T) {
	cases := []struct {
		name string
		src  domain.SourceRef
		want string
	}{
		{"manual default", domain.SourceRef{Type: "manual", ID: "default"}, "Prestation"},
		{"manual named", domain.SourceRef{Type: "manual", ID: "astreinte"}, "Prestation (astreinte)"},
		{"holiday", domain.SourceRef{Type: "holiday", ID: "2026-09-21"}, "Jour férié"},
		{"leave", domain.SourceRef{Type: "leave", ID: "abc"}, "Congé"},
		{"mission", domain.SourceRef{Type: "mission", ID: "abc"}, "Mission"},
		{"unknown type", domain.SourceRef{Type: "ett", ID: "abc"}, "Prestation"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sourceLabel(tc.src); got != tc.want {
				t.Fatalf("sourceLabel(%+v) = %q, want %q", tc.src, got, tc.want)
			}
		})
	}
}

func TestBuildLines_ColumnsMirrorTheGrid(t *testing.T) {
	demandID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	reader := &fakeWorkRefReader{subjects: map[uuid.UUID]string{demandID: "Refonte export"}}
	r := &TenantRenderer{workRefs: map[string]ports.WorkRefLabelReader{"tma": reader}}

	ts := domain.Timesheet{
		TenantID: testTenant,
		Weeks: []domain.WeekEntry{{Lines: []domain.TimeLine{
			{
				Day:         time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
				Duration:    kernel.Duration{Minutes: 375},
				Comment:     "Analyse",
				Source:      domain.SourceRef{Type: "manual", ID: "default"},
				WorkRefType: "tma",
				WorkRefID:   demandID.String(),
			},
			{
				Day:      time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
				Duration: kernel.Duration{Minutes: 480},
				Source:   domain.SourceRef{Type: "holiday", ID: "2026-09-02"},
			},
			{
				Day:      time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC),
				Duration: kernel.Duration{Minutes: 0},
				Source:   domain.SourceRef{Type: "manual", ID: "default"},
			},
		}}},
	}

	lines := r.buildLines(context.Background(), ts, "DURLET Pascal")
	if len(lines) != 2 {
		t.Fatalf("expected empty lines to be dropped, got %d lines", len(lines))
	}
	want := CRALine{
		Day:        "01/09/2026",
		Person:     "DURLET Pascal",
		Assignment: "Évolutions & incidents — Refonte export",
		Comment:    "Analyse",
		Duration:   "6 h 15",
	}
	if lines[0] != want {
		t.Fatalf("line[0] = %+v, want %+v", lines[0], want)
	}
	if lines[1].Assignment != "Jour férié" {
		t.Fatalf("absence assignment = %q, want %q", lines[1].Assignment, "Jour férié")
	}
}

func TestBuildLines_MemoizesWorkRefSubjects(t *testing.T) {
	demandID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	reader := &fakeWorkRefReader{subjects: map[uuid.UUID]string{demandID: "Hotline"}}
	r := &TenantRenderer{workRefs: map[string]ports.WorkRefLabelReader{"ticket": reader}}

	line := domain.TimeLine{
		Day:      time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		Duration: kernel.Duration{Minutes: 60},
		// A line fed by the support module carries the ticket as its source, with
		// no separate work ref — the grid hides the select in that case.
		Source: domain.SourceRef{Type: "ticket", ID: demandID.String()},
	}
	ts := domain.Timesheet{Weeks: []domain.WeekEntry{{Lines: []domain.TimeLine{line, line, line}}}}

	lines := r.buildLines(context.Background(), ts, "")
	for i, got := range lines {
		if got.Assignment != "Service utilisateur — Hotline" {
			t.Fatalf("line[%d].Assignment = %q", i, got.Assignment)
		}
	}
	if reader.calls != 1 {
		t.Fatalf("expected the subject to be read once, got %d reads", reader.calls)
	}
}

func TestAssignmentLabel_FallsBackWhenSubjectUnavailable(t *testing.T) {
	refID := "44444444-4444-4444-4444-444444444444"
	cases := []struct {
		name   string
		reader *fakeWorkRefReader
		want   string
	}{
		{"read error", &fakeWorkRefReader{err: errors.New("boom")}, "Exploitation & travaux #44444444"},
		{"empty subject", &fakeWorkRefReader{subjects: map[uuid.UUID]string{}}, "Exploitation & travaux #44444444"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &TenantRenderer{workRefs: map[string]ports.WorkRefLabelReader{"work_request": tc.reader}}
			got := r.assignmentLabel(context.Background(), testTenant, domain.TimeLine{
				Source:      domain.SourceRef{Type: "manual", ID: "default"},
				WorkRefType: "work_request",
				WorkRefID:   refID,
			}, map[string]string{})
			if got != tc.want {
				t.Fatalf("assignmentLabel = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCollaborateur_TolerantToDirectoryFailure(t *testing.T) {
	r := &TenantRenderer{users: fakeUserResolver{err: errors.New("boom")}}
	if got := r.collaborateur(context.Background(), domain.Timesheet{}); got != "" {
		t.Fatalf("collaborateur = %q, want empty on resolver error", got)
	}

	r = &TenantRenderer{users: fakeUserResolver{identity: ports.UserIdentity{Nom: "DURLET", Prenom: "Pascal"}}}
	if got := r.collaborateur(context.Background(), domain.Timesheet{}); got != "DURLET Pascal" {
		t.Fatalf("collaborateur = %q, want %q", got, "DURLET Pascal")
	}
}
