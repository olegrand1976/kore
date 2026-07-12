package pdf

import (
	"context"

	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
)

// TenantRenderer loads société branding before rendering the CRA document.
type TenantRenderer struct {
	org   orgports.OrganizationService
	inner *HTMLRenderer
}

func NewTenantRenderer(org orgports.OrganizationService) ports.PDFRenderer {
	return &TenantRenderer{org: org, inner: NewHTMLRenderer()}
}

func (r *TenantRenderer) Render(ctx context.Context, ts domain.Timesheet) (domain.Document, error) {
	brand := CRABrandData{
		CompanyName:    "Kore",
		ShowKoreFooter: true,
	}
	societes, err := r.org.ListSocietes(ctx, ts.TenantID)
	if err == nil && len(societes) > 0 {
		s := societes[0]
		if s.RaisonSociale != "" {
			brand.CompanyName = s.RaisonSociale
		}
		brand.CompanyAddr = trimJoin(s.Adresse, s.URLTenant)
		brand.CompanyLogo = s.Logo
	}

	content, err := r.inner.Render(ctx, ts, brand)
	if err != nil {
		return domain.Document{}, err
	}
	return domain.Document{
		Filename: r.inner.Filename(ts),
		Content:  content,
		MimeType: "text/html",
	}, nil
}

func trimJoin(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return ""
	}
	result := out[0]
	for i := 1; i < len(out); i++ {
		result += " · " + out[i]
	}
	return result
}

var _ ports.PDFRenderer = (*TenantRenderer)(nil)
