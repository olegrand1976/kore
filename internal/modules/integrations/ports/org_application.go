package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

type CreateApplicationInput struct {
	TenantID           kernel.TenantID
	Libelle            string
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

type ApplicationSummary struct {
	ID      uuid.UUID
	Libelle string
}

type ApplicationCreator interface {
	CreateApplication(ctx context.Context, cmd CreateApplicationInput) (ApplicationSummary, error)
	DeactivateApplication(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) error
}
