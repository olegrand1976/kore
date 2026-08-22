package http

import (
	"context"

	integrationports "github.com/kore/kore/internal/modules/integrations/ports"
)

// TaigaApplicationBridge links Kore applications to Taiga projects (wired from integrations).
type TaigaApplicationBridge interface {
	CreateApplicationWithTaiga(ctx context.Context, cmd integrationports.CreateApplicationInput, taigaProjectID int) (integrationports.ApplicationSummary, error)
}
