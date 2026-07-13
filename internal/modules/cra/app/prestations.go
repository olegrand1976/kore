package app

import (
	"context"
	"strings"

	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

func (s *Service) ListPrestations(ctx context.Context, tenant kernel.TenantID, month domain.Month) ([]domain.TimesheetSummary, error) {
	return s.repo.ListSummariesByTenantMonth(ctx, tenant, month)
}

func (s *Service) RejectTimesheet(ctx context.Context, cmd ports.RejectTimesheetCommand) error {
	ts, err := s.repo.GetByID(ctx, cmd.TenantID, cmd.TimesheetID)
	if err != nil {
		return err
	}
	reason := strings.TrimSpace(cmd.Reason)
	if reason == "" {
		reason = "Rejet manager"
	}
	if err := ts.Reject(s.clock.Now(), cmd.ManagerID, reason); err != nil {
		return err
	}
	return s.repo.Save(ctx, ts)
}

func (s *Service) ValidateAll(ctx context.Context, cmd ports.ValidateAllCommand) (int, error) {
	summaries, err := s.ListPrestations(ctx, cmd.TenantID, cmd.Month)
	if err != nil {
		return 0, err
	}
	validated := 0
	for _, summary := range summaries {
		if summary.Status == domain.StatusDefinitif {
			continue
		}
		if summary.Status != domain.StatusValideSemaine {
			continue
		}
		if err := s.ValidateFinal(ctx, ports.ManagerValidateCommand{
			TenantID:    cmd.TenantID,
			TimesheetID: summary.ID,
			ManagerID:   cmd.ManagerID,
		}); err != nil {
			continue
		}
		validated++
	}
	return validated, nil
}
