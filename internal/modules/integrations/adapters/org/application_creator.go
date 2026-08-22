package org

import (
	"context"

	"github.com/google/uuid"
	integrationports "github.com/kore/kore/internal/modules/integrations/ports"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type ApplicationCreator struct {
	org orgports.OrganizationService
}

func NewApplicationCreator(org orgports.OrganizationService) *ApplicationCreator {
	return &ApplicationCreator{org: org}
}

func (a *ApplicationCreator) CreateApplication(ctx context.Context, cmd integrationports.CreateApplicationInput) (integrationports.ApplicationSummary, error) {
	app, err := a.org.CreateApplication(ctx, orgports.CreateApplicationCommand{
		TenantID:           cmd.TenantID,
		Libelle:            cmd.Libelle,
		Proprietaire:       cmd.Proprietaire,
		ModeFacturation:    cmd.ModeFacturation,
		UOActivee:          cmd.UOActivee,
		ChefUtilisateurID:  cmd.ChefUtilisateurID,
		DefaultTJMCents:    cmd.DefaultTJMCents,
		SiteIDs:            cmd.SiteIDs,
		ServiceIDs:         cmd.ServiceIDs,
		EquipeIDs:          cmd.EquipeIDs,
		MethodologyProfile: cmd.MethodologyProfile,
	})
	if err != nil {
		return integrationports.ApplicationSummary{}, err
	}
	return integrationports.ApplicationSummary{ID: app.ID, Libelle: app.Libelle}, nil
}

func (a *ApplicationCreator) DeactivateApplication(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) error {
	_, err := a.org.SetApplicationActive(ctx, orgports.SetApplicationActiveCommand{
		TenantID:      tenant,
		ApplicationID: applicationID,
		Active:        false,
	})
	return err
}
