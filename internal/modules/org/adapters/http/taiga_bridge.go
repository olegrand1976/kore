package http

import (
	"context"

	"github.com/google/uuid"
	integrationports "github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

// TaigaApplicationBridge links Kore applications to Taiga projects (wired from integrations).
type TaigaApplicationBridge interface {
	CreateApplicationWithTaiga(ctx context.Context, cmd integrationports.CreateApplicationInput, taigaProjectID int) (integrationports.ApplicationSummary, error)
	LinkExistingApplication(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID, taigaProjectID int) error
}
