package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/support/domain"
	"github.com/kore/kore/pkg/kernel"
)

func (s *service) IngestInboundEmails(ctx context.Context, tenant kernel.TenantID, applicationID uuid.UUID) (int, error) {
	if s.mail == nil {
		return 0, nil
	}
	emails, err := s.mail.Poll(ctx)
	if err != nil {
		return 0, err
	}
	created := 0
	for _, email := range emails {
		t := domain.NewTicket(tenant, applicationID, email.Subject, email.Body, email.ReporterID)
		t.Channel = "mail"
		if err := s.repo.SaveTicket(ctx, t); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}
