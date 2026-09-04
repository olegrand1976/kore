package pdf

import (
	"context"
	"strconv"
	"strings"

	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
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
		if addr := orgdomain.FormatSocieteAddress(s); addr != "" {
			brand.CompanyAddr = addr
		} else {
			brand.CompanyAddr = trimJoin(s.Adresse, s.URLTenant)
		}
		brand.CompanyLogo = s.Logo
	}

	brand.Lines = buildLinesFromTimesheet(ts)
	info := ts.CommercialInfo
	brand.Client = info.Client
	brand.Mission = info.Mission
	brand.Description = info.Description
	brand.Lieu = info.Lieu
	brand.ResponsableClient = info.ResponsableClient
	if len(info.Technologies) > 0 {
		brand.Technologies = strings.Join(info.Technologies, ", ")
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

func buildLinesFromTimesheet(ts domain.Timesheet) []CRALine {
	var out []CRALine
	for _, week := range ts.Weeks {
		for _, line := range week.Lines {
			if line.Duration.Minutes <= 0 {
				continue
			}
			out = append(out, CRALine{
				Task:  line.Source.Type + "/" + line.Source.ID,
				Days:  line.Day.Format("02/01/2006"),
				Hours: formatHours(line.Duration.Minutes),
			})
		}
	}
	return out
}

// formatHours renders a duration in hours with up to 2 decimals and no trailing
// zeros: 480 -> "8", 450 -> "7.5", 375 -> "6.25". Formatting straight from
// minutes keeps quarter-hours exact and matches the web grid, where "%.1f" used
// to round 6.25 down to "6.2" (ties-to-even) while the UI showed "6.3".
func formatHours(minutes int) string {
	if minutes%60 == 0 {
		return strconv.Itoa(minutes / 60)
	}
	s := strconv.FormatFloat(float64(minutes)/60, 'f', 2, 64)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
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
