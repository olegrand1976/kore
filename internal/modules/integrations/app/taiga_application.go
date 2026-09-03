package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ImportApplicationProject struct {
	TaigaProjectID int    `json:"taigaProjectId"`
	Libelle        string `json:"libelle,omitempty"`
}

type ImportApplicationsCommand struct {
	TenantID           kernel.TenantID
	Projects           []ImportApplicationProject
	Proprietaire       string
	ModeFacturation    string
	UOActivee          bool
	ChefUtilisateurID  *uuid.UUID
	DefaultTJMCents    int64
	SiteIDs            []uuid.UUID
	ServiceIDs         []uuid.UUID
	EquipeIDs          []uuid.UUID
	MethodologyProfile string
}

type ImportApplicationError struct {
	TaigaProjectID int    `json:"taigaProjectId"`
	Message        string `json:"message"`
}

type ImportApplicationsResult struct {
	Created []struct {
		ApplicationID  uuid.UUID `json:"applicationId"`
		Libelle        string    `json:"libelle"`
		TaigaProjectID int       `json:"taigaProjectId"`
	} `json:"created"`
	Errors []ImportApplicationError `json:"errors"`
}

func (s *TaigaService) ListUnlinkedProjects(ctx context.Context, tenant kernel.TenantID) ([]ports.TaigaProject, error) {
	if s.gateway == nil {
		return nil, domain.ErrTaigaNotConfigured
	}
	projects, err := s.gateway.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	linked, err := s.repo.ListLinkedTaigaProjectIDs(ctx, tenant)
	if err != nil {
		return nil, err
	}
	linkedSet := make(map[string]struct{}, len(linked))
	for _, id := range linked {
		linkedSet[id] = struct{}{}
	}
	out := make([]ports.TaigaProject, 0, len(projects))
	for _, p := range projects {
		if _, ok := linkedSet[strconv.Itoa(p.ID)]; ok {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func (s *TaigaService) FindByKoreApplication(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) (domain.ExternalLink, error) {
	return s.repo.FindExternalLinkByKore(ctx, tenant, "application", applicationID)
}

func (s *TaigaService) ListLinkedApplicationIDs(ctx context.Context, tenant kernel.TenantID) ([]uuid.UUID, error) {
	return s.repo.ListLinkedApplicationIDs(ctx, tenant)
}

func (s *TaigaService) LinkExistingApplication(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID, taigaProjectID int) error {
	if taigaProjectID <= 0 {
		return domain.ErrTaigaProjectNotFound
	}
	_, err := s.FindByKoreApplication(ctx, tenant, applicationID)
	if err == nil {
		return domain.ErrTaigaApplicationAlreadyLinked
	}
	if !errors.Is(err, domain.ErrExternalLinkNotFound) {
		return err
	}
	project, err := s.findTaigaProject(ctx, taigaProjectID)
	if err != nil {
		return err
	}
	return s.LinkApplicationToProject(ctx, tenant, applicationID, taigaProjectID, project)
}

func (s *TaigaService) LinkApplicationToProject(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID, taigaProjectID int, project ports.TaigaProject) error {
	if taigaProjectID <= 0 {
		return domain.ErrTaigaProjectNotFound
	}
	if err := s.ensureTaigaProjectAvailable(ctx, tenant, taigaProjectID); err != nil {
		return err
	}
	if project.ID == 0 {
		project.ID = taigaProjectID
	}
	url := taigaProjectURL(s.cfg.BaseURL, project.Slug)
	now := time.Now().UTC()
	pid := project.ID
	link := domain.ExternalLink{
		TenantID:          tenant,
		Provider:          "taiga",
		ExternalType:      "project",
		ExternalID:        strconv.Itoa(taigaProjectID),
		ExternalProjectID: &pid,
		ExternalURL:       url,
		KoreEntityType:    "application",
		KoreEntityID:      applicationID,
		Metadata:          map[string]any{"slug": project.Slug, "name": project.Name},
		LastSyncAt:        &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return s.repo.InsertApplicationProjectLink(ctx, link)
}

func (s *TaigaService) CreateApplicationWithTaiga(ctx context.Context, cmd ports.CreateApplicationInput, taigaProjectID int) (ports.ApplicationSummary, error) {
	if s.apps == nil {
		return ports.ApplicationSummary{}, fmt.Errorf("application creator not configured")
	}
	if taigaProjectID <= 0 {
		return ports.ApplicationSummary{}, domain.ErrTaigaProjectNotFound
	}
	project, err := s.findTaigaProject(ctx, taigaProjectID)
	if err != nil {
		return ports.ApplicationSummary{}, err
	}
	return s.createApplicationWithTaigaProject(ctx, cmd, taigaProjectID, project)
}

func (s *TaigaService) ImportApplications(ctx context.Context, cmd ImportApplicationsCommand) (ImportApplicationsResult, error) {
	if s.apps == nil {
		return ImportApplicationsResult{}, fmt.Errorf("application creator not configured")
	}
	if len(cmd.Projects) == 0 {
		return ImportApplicationsResult{}, fmt.Errorf("projects required")
	}
	if s.gateway == nil {
		return ImportApplicationsResult{}, domain.ErrTaigaNotConfigured
	}
	allProjects, err := s.gateway.ListProjects(ctx)
	if err != nil {
		return ImportApplicationsResult{}, err
	}
	projectByID := make(map[int]ports.TaigaProject, len(allProjects))
	for _, p := range allProjects {
		projectByID[p.ID] = p
	}

	result := ImportApplicationsResult{
		Created: make([]struct {
			ApplicationID  uuid.UUID `json:"applicationId"`
			Libelle        string    `json:"libelle"`
			TaigaProjectID int       `json:"taigaProjectId"`
		}, 0),
		Errors: make([]ImportApplicationError, 0),
	}
	base := ports.CreateApplicationInput{
		TenantID:           cmd.TenantID,
		Proprietaire:       cmd.Proprietaire,
		ModeFacturation:    cmd.ModeFacturation,
		UOActivee:          cmd.UOActivee,
		ChefUtilisateurID:  cmd.ChefUtilisateurID,
		DefaultTJMCents:    cmd.DefaultTJMCents,
		SiteIDs:            cmd.SiteIDs,
		ServiceIDs:         cmd.ServiceIDs,
		EquipeIDs:          cmd.EquipeIDs,
		MethodologyProfile: cmd.MethodologyProfile,
	}
	for _, item := range cmd.Projects {
		project, ok := projectByID[item.TaigaProjectID]
		if !ok {
			result.Errors = append(result.Errors, ImportApplicationError{
				TaigaProjectID: item.TaigaProjectID,
				Message:        domain.ErrTaigaProjectNotFound.Error(),
			})
			continue
		}
		in := base
		in.Libelle = strings.TrimSpace(item.Libelle)
		app, err := s.createApplicationWithTaigaProject(ctx, in, item.TaigaProjectID, project)
		if err != nil {
			result.Errors = append(result.Errors, ImportApplicationError{
				TaigaProjectID: item.TaigaProjectID,
				Message:        err.Error(),
			})
			continue
		}
		result.Created = append(result.Created, struct {
			ApplicationID  uuid.UUID `json:"applicationId"`
			Libelle        string    `json:"libelle"`
			TaigaProjectID int       `json:"taigaProjectId"`
		}{
			ApplicationID:  app.ID,
			Libelle:        app.Libelle,
			TaigaProjectID: item.TaigaProjectID,
		})
	}
	return result, nil
}

func (s *TaigaService) createApplicationWithTaigaProject(
	ctx context.Context,
	cmd ports.CreateApplicationInput,
	taigaProjectID int,
	project ports.TaigaProject,
) (ports.ApplicationSummary, error) {
	if err := s.ensureTaigaProjectAvailable(ctx, cmd.TenantID, taigaProjectID); err != nil {
		return ports.ApplicationSummary{}, err
	}
	if strings.TrimSpace(cmd.Libelle) == "" {
		cmd.Libelle = project.Name
	}
	app, err := s.apps.CreateApplication(ctx, cmd)
	if err != nil {
		return ports.ApplicationSummary{}, err
	}
	if err := s.LinkApplicationToProject(ctx, cmd.TenantID, app.ID, taigaProjectID, project); err != nil {
		_ = s.apps.DeactivateApplication(ctx, cmd.TenantID, app.ID)
		return ports.ApplicationSummary{}, err
	}
	return app, nil
}

func (s *TaigaService) ensureTaigaProjectAvailable(ctx context.Context, tenant kernel.TenantID, taigaProjectID int) error {
	linked, err := s.repo.ListLinkedTaigaProjectIDs(ctx, tenant)
	if err != nil {
		return err
	}
	extID := strconv.Itoa(taigaProjectID)
	if slices.Contains(linked, extID) {
		return domain.ErrTaigaProjectLinked
	}
	return nil
}

func (s *TaigaService) findTaigaProject(ctx context.Context, taigaProjectID int) (ports.TaigaProject, error) {
	if s.gateway == nil {
		return ports.TaigaProject{}, domain.ErrTaigaNotConfigured
	}
	projects, err := s.gateway.ListProjects(ctx)
	if err != nil {
		return ports.TaigaProject{}, err
	}
	for _, p := range projects {
		if p.ID == taigaProjectID {
			return p, nil
		}
	}
	return ports.TaigaProject{}, domain.ErrTaigaProjectNotFound
}

func taigaProjectURL(baseURL, slug string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	slug = strings.TrimSpace(slug)
	if base == "" || slug == "" {
		return ""
	}
	return fmt.Sprintf("%s/project/%s", base, slug)
}
