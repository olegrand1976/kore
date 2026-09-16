package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

const (
	syncOriginKore    = "kore"
	syncOriginTaiga   = "taiga"
	externalTypeIssue = "issue"
)

// TaigaSyncResult summarizes a bidirectional sync run.
type TaigaSyncResult struct {
	Pulled  int      `json:"pulled"`
	Pushed  int      `json:"pushed"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors"`
}

func (s *TaigaService) resolveSyncAuthor(ctx context.Context, tenant kernel.TenantID, preferred *uuid.UUID) (uuid.UUID, error) {
	if preferred != nil && *preferred != uuid.Nil {
		return *preferred, nil
	}
	mappings, err := s.repo.ListUserMappings(ctx, tenant, "taiga")
	if err != nil {
		return uuid.Nil, err
	}
	if len(mappings) == 0 {
		return uuid.Nil, domain.ErrTaigaSyncAuthorRequired
	}
	return mappings[0].KoreUserID, nil
}

func koreExternalRef(demandID uuid.UUID) []string {
	return []string{"kore", demandID.String()}
}

func parseKoreDemandRef(ext any) (uuid.UUID, bool) {
	extSlice, ok := ext.([]any)
	if !ok || len(extSlice) < 2 {
		return uuid.Nil, false
	}
	if fmt.Sprint(extSlice[0]) != "kore" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(fmt.Sprint(extSlice[1]))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func isKoreRefShape(ext any) bool {
	extSlice, ok := ext.([]any)
	if !ok || len(extSlice) < 2 {
		return false
	}
	return fmt.Sprint(extSlice[0]) == "kore"
}

func hasKoreExternalRefStrings(refs []string) bool {
	return len(refs) >= 2 && refs[0] == "kore"
}

func (s *TaigaService) upsertDemandIssueLink(
	ctx context.Context,
	tenant kernel.TenantID,
	demandID uuid.UUID,
	issue ports.TaigaIssue,
	origin string,
) error {
	now := time.Now().UTC()
	projectID := issue.ProjectID
	ref := issue.Ref
	url := strings.TrimSpace(issue.Permalink)
	if url == "" {
		base := strings.TrimRight(strings.TrimSpace(s.cfg.BaseURL), "/")
		slug := strings.TrimSpace(s.cfg.ProjectSlug)
		if base != "" && slug != "" && issue.Ref > 0 {
			url = fmt.Sprintf("%s/project/%s/issue/%d", base, slug, issue.Ref)
		}
	}
	link := domain.ExternalLink{
		TenantID:          tenant,
		Provider:          "taiga",
		ExternalType:      externalTypeIssue,
		ExternalID:        strconv.Itoa(issue.ID),
		ExternalProjectID: &projectID,
		ExternalRef:       &ref,
		ExternalURL:       url,
		KoreEntityType:    "demand",
		KoreEntityID:      demandID,
		Metadata:          map[string]any{"syncOrigin": origin},
		LastSyncAt:        &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return s.repo.UpsertExternalLink(ctx, link)
}

// PullIssueToKore creates a TMA demand from a Taiga issue and links them.
func (s *TaigaService) PullIssueToKore(
	ctx context.Context,
	tenant kernel.TenantID,
	issue ports.TaigaIssue,
	authorID uuid.UUID,
) (uuid.UUID, error) {
	if s.demands == nil {
		return uuid.Nil, domain.ErrTaigaNotConfigured
	}
	if issue.ID <= 0 || issue.ProjectID <= 0 {
		return uuid.Nil, domain.ErrTaigaProjectNotFound
	}
	if _, err := s.repo.FindExternalLinkByExternal(ctx, tenant, "taiga", externalTypeIssue, strconv.Itoa(issue.ID)); err == nil {
		return uuid.Nil, domain.ErrTaigaIssueAlreadyLinked
	} else if err != domain.ErrExternalLinkNotFound {
		return uuid.Nil, err
	}

	appLink, err := s.repo.FindApplicationByTaigaProjectID(ctx, tenant, strconv.Itoa(issue.ProjectID))
	if err != nil {
		return uuid.Nil, err
	}
	subject := strings.TrimSpace(issue.Subject)
	if subject == "" {
		subject = fmt.Sprintf("Taiga issue #%d", issue.Ref)
	}
	demandID, err := s.demands.CreateDemandFromTaiga(
		ctx, tenant, appLink.KoreEntityID, authorID, subject, strings.TrimSpace(issue.Description),
	)
	if err != nil {
		return uuid.Nil, err
	}
	if err := s.upsertDemandIssueLink(ctx, tenant, demandID, issue, syncOriginTaiga); err != nil {
		return uuid.Nil, err
	}
	if s.gateway != nil && !hasKoreExternalRefStrings(issue.ExternalReference) {
		if _, err := s.gateway.UpdateIssueExternalReference(ctx, issue.ID, issue.Version, koreExternalRef(demandID)); err != nil {
			slog.Warn("taiga set external_reference failed",
				"issueId", issue.ID, "demandId", demandID.String(), "error", err)
		}
	}
	return demandID, nil
}

// PushDemandToTaiga creates a Taiga issue for a Kore TMA demand when the application is linked.
func (s *TaigaService) PushDemandToTaiga(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.ExternalLink, error) {
	if s.gateway == nil {
		return domain.ExternalLink{}, domain.ErrTaigaNotConfigured
	}
	if s.demands == nil {
		return domain.ExternalLink{}, domain.ErrTaigaNotConfigured
	}
	if _, err := s.repo.FindExternalLinkByKore(ctx, tenant, "demand", demandID); err == nil {
		return domain.ExternalLink{}, domain.ErrTaigaDemandAlreadyLinked
	} else if err != domain.ErrExternalLinkNotFound {
		return domain.ExternalLink{}, err
	}
	summary, err := s.demands.GetDemandSummary(ctx, tenant, demandID)
	if err != nil {
		return domain.ExternalLink{}, err
	}
	appLink, err := s.repo.FindExternalLinkByKore(ctx, tenant, "application", summary.ApplicationID)
	if err != nil {
		if err == domain.ErrExternalLinkNotFound {
			return domain.ExternalLink{}, domain.ErrTaigaApplicationNotLinked
		}
		return domain.ExternalLink{}, err
	}
	projectID, err := strconv.Atoi(appLink.ExternalID)
	if err != nil || projectID <= 0 {
		return domain.ExternalLink{}, domain.ErrTaigaProjectNotFound
	}
	issue, err := s.gateway.CreateIssue(ctx, projectID, summary.Subject, summary.Description, koreExternalRef(demandID))
	if err != nil {
		return domain.ExternalLink{}, err
	}
	if err := s.upsertDemandIssueLink(ctx, tenant, demandID, issue, syncOriginKore); err != nil {
		return domain.ExternalLink{}, err
	}
	return s.repo.FindExternalLinkByKore(ctx, tenant, "demand", demandID)
}

// OnDemandCreated is the TMA lifecycle hook: best-effort push when the app is linked to Taiga.
func (s *TaigaService) OnDemandCreated(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) error {
	if s == nil || s.gateway == nil || s.demands == nil {
		return nil
	}
	summary, err := s.demands.GetDemandSummary(ctx, tenant, demandID)
	if err != nil {
		return nil
	}
	if _, err := s.repo.FindExternalLinkByKore(ctx, tenant, "application", summary.ApplicationID); err != nil {
		return nil
	}
	if _, err := s.PushDemandToTaiga(ctx, tenant, demandID); err != nil {
		slog.Warn("taiga auto-push on demand create failed",
			"demandId", demandID.String(), "tenantId", tenant.String(), "error", err)
	}
	return nil
}

// SyncLinkedProjects pulls unlinked Taiga issues and pushes unlinked Kore demands for all app↔project links.
func (s *TaigaService) SyncLinkedProjects(
	ctx context.Context,
	tenant kernel.TenantID,
	actorID *uuid.UUID,
) (TaigaSyncResult, error) {
	result := TaigaSyncResult{Errors: []string{}}
	if s.gateway == nil {
		return result, domain.ErrTaigaNotConfigured
	}
	authorID, err := s.resolveSyncAuthor(ctx, tenant, actorID)
	if err != nil {
		return result, err
	}
	links, err := s.repo.ListApplicationProjectLinks(ctx, tenant)
	if err != nil {
		return result, err
	}
	for _, appLink := range links {
		projectID, err := strconv.Atoi(appLink.ExternalID)
		if err != nil || projectID <= 0 {
			result.Skipped++
			continue
		}
		issues, err := s.gateway.ListProjectIssues(ctx, projectID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("list issues project %d: %v", projectID, err))
			continue
		}
		for _, issue := range issues {
			if _, err := s.repo.FindExternalLinkByExternal(ctx, tenant, "taiga", externalTypeIssue, strconv.Itoa(issue.ID)); err == nil {
				result.Skipped++
				continue
			} else if err != domain.ErrExternalLinkNotFound {
				result.Errors = append(result.Errors, err.Error())
				continue
			}
			if hasKoreExternalRefStrings(issue.ExternalReference) {
				// Already stamped from Kore but link missing — recover link only if demand exists.
				demandID, parseErr := uuid.Parse(issue.ExternalReference[1])
				if parseErr == nil {
					if exists, _ := s.demands.KoreDemandExists(ctx, tenant, demandID); exists {
						_ = s.upsertDemandIssueLink(ctx, tenant, demandID, issue, syncOriginKore)
						result.Skipped++
						continue
					}
				}
			}
			if _, err := s.PullIssueToKore(ctx, tenant, issue, authorID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("pull issue %d: %v", issue.ID, err))
				continue
			}
			result.Pulled++
		}

		demands, err := s.demands.ListDemandsByApplication(ctx, tenant, appLink.KoreEntityID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("list demands app %s: %v", appLink.KoreEntityID, err))
			continue
		}
		for _, d := range demands {
			if _, err := s.repo.FindExternalLinkByKore(ctx, tenant, "demand", d.ID); err == nil {
				result.Skipped++
				continue
			} else if err != domain.ErrExternalLinkNotFound {
				result.Errors = append(result.Errors, err.Error())
				continue
			}
			if _, err := s.PushDemandToTaiga(ctx, tenant, d.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("push demand %s: %v", d.ID, err))
				continue
			}
			result.Pushed++
		}
	}
	return result, nil
}
