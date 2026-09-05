package pdf

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"net"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

// TenantRenderer loads société branding before rendering the CRA document.
type TenantRenderer struct {
	org         orgports.OrganizationService
	users       ports.UserIdentityResolver
	workRefs    map[string]ports.WorkRefLabelReader
	productSite string
	inner       *HTMLRenderer
}

type TenantRendererOption func(*TenantRenderer)

// WithUserIdentityResolver names the CRA owner on the document; without it the
// "Personne" column stays empty rather than showing a raw UUID.
func WithUserIdentityResolver(users ports.UserIdentityResolver) TenantRendererOption {
	return func(r *TenantRenderer) { r.users = users }
}

// WithWorkRefLabelReaders resolves the subject of the work item each line is
// charged to, keyed by work ref type ("tma", "ticket", "work_request").
func WithWorkRefLabelReaders(readers map[string]ports.WorkRefLabelReader) TenantRendererOption {
	return func(r *TenantRenderer) { r.workRefs = readers }
}

// WithProductSite names the Kore site in the product footer. Takes the public
// frontend origin and keeps only its host; a dev origin (localhost, bare IP) is
// dropped rather than printed on a document meant to be sent to a client.
func WithProductSite(rawURL string) TenantRendererOption {
	return func(r *TenantRenderer) { r.productSite = productSiteHost(rawURL) }
}

func productSiteHost(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	host := u.Hostname()
	if host == "" || host == "localhost" || net.ParseIP(host) != nil || !strings.Contains(host, ".") {
		return ""
	}
	return host
}

func NewTenantRenderer(org orgports.OrganizationService, opts ...TenantRendererOption) ports.PDFRenderer {
	r := &TenantRenderer{org: org, inner: NewHTMLRenderer()}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *TenantRenderer) Render(ctx context.Context, ts domain.Timesheet) (domain.Document, error) {
	brand := CRABrandData{
		CompanyName:    "Kore",
		ShowKoreFooter: true,
		ProductSite:    r.productSite,
	}
	declaredLogo := ""
	societes, err := r.societes(ctx, ts.TenantID)
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
		declaredLogo = s.Logo
	}
	brand.CompanyLogo = r.companyLogo(ctx, ts.TenantID, declaredLogo)

	brand.Collaborateur = r.collaborateur(ctx, ts)
	brand.Lines = r.buildLines(ctx, ts, brand.Collaborateur)
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

func (r *TenantRenderer) societes(ctx context.Context, tenant kernel.TenantID) ([]orgdomain.Societe, error) {
	if r.org == nil {
		return nil, nil
	}
	return r.org.ListSocietes(ctx, tenant)
}

// companyLogo embeds the société logo as a data URI. The PDF is printed from a
// data: page, where the stored logo's own API path would not resolve and an
// external fetch would make the export depend on the network; only a logo
// declared as an absolute URL is left to the browser to fetch.
func (r *TenantRenderer) companyLogo(ctx context.Context, tenant kernel.TenantID, declared string) template.URL {
	if r.org != nil {
		content, contentType, err := r.org.GetTenantLogo(ctx, tenant)
		if err == nil && len(content) > 0 {
			if mime, ok := imageMIME(contentType); ok {
				return template.URL("data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(content))
			}
		}
	}
	if strings.HasPrefix(declared, "https://") || strings.HasPrefix(declared, "http://") {
		return template.URL(declared)
	}
	return ""
}

// imageMIME keeps the data URI to image types a print engine renders — an
// unknown content type is a corrupt upload, not a logo worth risking in an
// unfiltered attribute.
func imageMIME(contentType string) (string, bool) {
	mime := strings.ToLower(strings.TrimSpace(contentType))
	if idx := strings.Index(mime, ";"); idx >= 0 {
		mime = strings.TrimSpace(mime[:idx])
	}
	switch mime {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/svg+xml":
		return mime, true
	default:
		return "", false
	}
}

// collaborateur formats the CRA owner as "Nom Prénom", falling back to the login
// when the directory has no civil name. A failed lookup is not fatal: the
// document is still worth producing without the name.
func (r *TenantRenderer) collaborateur(ctx context.Context, ts domain.Timesheet) string {
	if r.users == nil {
		return ""
	}
	identity, err := r.users.ResolveUserIdentity(ctx, ts.TenantID, ts.UserID)
	if err != nil {
		return ""
	}
	return formatPerson(identity)
}

func formatPerson(identity ports.UserIdentity) string {
	name := strings.TrimSpace(strings.TrimSpace(identity.Nom) + " " + strings.TrimSpace(identity.Prenom))
	if name != "" {
		return name
	}
	return strings.TrimSpace(identity.Login)
}

func (r *TenantRenderer) buildLines(ctx context.Context, ts domain.Timesheet, person string) []CRALine {
	var out []CRALine
	labels := map[string]string{}
	for _, week := range ts.Weeks {
		for _, line := range week.Lines {
			if line.Duration.Minutes <= 0 {
				continue
			}
			out = append(out, CRALine{
				Day:        line.Day.Format("02/01/2006"),
				Person:     person,
				Assignment: r.assignmentLabel(ctx, ts.TenantID, line, labels),
				Comment:    line.Comment,
				Duration:   formatDuration(line.Duration.Minutes),
			})
		}
	}
	return out
}

// assignmentLabel mirrors the CRA grid: a line either points at a work item via
// its work ref, or is itself fed by one (source type tma/ticket/work_request);
// anything else — manual entry, leave, public holiday — falls back to its source
// label. Subjects are memoized: the same demand usually spans many lines.
func (r *TenantRenderer) assignmentLabel(ctx context.Context, tenant kernel.TenantID, line domain.TimeLine, labels map[string]string) string {
	refType, refID := line.WorkRefType, line.WorkRefID
	if refType == "" {
		if _, ok := r.workRefs[line.Source.Type]; !ok {
			return sourceLabel(line.Source)
		}
		refType, refID = line.Source.Type, line.Source.ID
	}

	key := refType + ":" + refID
	if label, ok := labels[key]; ok {
		return label
	}
	label := r.workRefLabel(ctx, tenant, refType, refID)
	labels[key] = label
	return label
}

func (r *TenantRenderer) workRefLabel(ctx context.Context, tenant kernel.TenantID, refType, refID string) string {
	fallback := workRefTypeLabel(refType)
	if refID != "" {
		fallback += " #" + shortID(refID)
	}
	reader, ok := r.workRefs[refType]
	if !ok {
		return fallback
	}
	id, err := uuid.Parse(refID)
	if err != nil {
		return fallback
	}
	subject, err := reader.WorkRefLabel(ctx, tenant, id)
	if err != nil || strings.TrimSpace(subject) == "" {
		return fallback
	}
	return workRefTypeLabel(refType) + " — " + subject
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// sourceLabel / workRefTypeLabel mirror frontend useCraSourceLabels and
// useCraWorkRefs so the PDF reads like the grid it exports.
func workRefTypeLabel(refType string) string {
	switch refType {
	case "tma":
		return "Évolutions & incidents"
	case "ticket":
		return "Service utilisateur"
	case "work_request":
		return "Exploitation & travaux"
	default:
		return "Prestation"
	}
}

func sourceLabel(src domain.SourceRef) string {
	switch src.Type {
	case "mission":
		return "Mission"
	case "leave":
		return "Congé"
	case "holiday":
		return "Jour férié"
	case "interne":
		return "Interne"
	case "formation":
		return "Formation"
	case "manual", "":
		if src.ID != "" && src.ID != "default" {
			return "Prestation (" + src.ID + ")"
		}
		return "Prestation"
	default:
		return "Prestation"
	}
}

// formatDuration renders minutes as hours and minutes — 375 -> "6 h 15",
// 480 -> "8 h", 45 -> "0 h 45" — keeping quarter-hours exact and the column
// aligned on the same unit.
func formatDuration(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}
	hours, mins := minutes/60, minutes%60
	if mins == 0 {
		return fmt.Sprintf("%d h", hours)
	}
	return fmt.Sprintf("%d h %02d", hours, mins)
}

func trimJoin(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " · ")
}

var _ ports.PDFRenderer = (*TenantRenderer)(nil)
