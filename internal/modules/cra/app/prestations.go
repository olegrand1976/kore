package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
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
	if err := s.repo.Save(ctx, ts); err != nil {
		return err
	}
	s.notifyTimesheetRejected(ctx, ts, reason)
	return nil
}

func (s *Service) notifyTimesheetRejected(ctx context.Context, ts domain.Timesheet, reason string) {
	if s.notifier == nil || s.emails == nil {
		return
	}
	recipients, err := s.emails.ResolveUserEmails(ctx, ts.TenantID, []uuid.UUID{ts.UserID})
	if err != nil || len(recipients) == 0 {
		return
	}
	subject := fmt.Sprintf("CRA %s rejeté", ts.Month)
	body := fmt.Sprintf(
		"Votre compte-rendu d'activité pour %s a été rejeté par votre manager.\n\nMotif : %s\n\nVous pouvez le corriger et le soumettre à nouveau.",
		ts.Month,
		reason,
	)
	_ = s.notifier.NotifyTransactional(ctx, notifports.TransactionalMessage{
		Subject:    subject,
		Body:       body,
		Recipients: recipients,
	})
}

func (s *Service) ValidateAll(ctx context.Context, cmd ports.ValidateAllCommand) (ports.ValidateAllResult, error) {
	summaries, err := s.ListPrestations(ctx, cmd.TenantID, cmd.Month)
	if err != nil {
		return ports.ValidateAllResult{}, err
	}
	result := ports.ValidateAllResult{}
	for _, summary := range summaries {
		if summary.Status == domain.StatusDefinitif {
			continue
		}
		if summary.Status != domain.StatusValideSemaine {
			result.Failed = append(result.Failed, ports.ValidateAllFailure{
				TimesheetID: summary.ID,
				Reason:      fmt.Sprintf("statut %s", summary.Status),
			})
			continue
		}
		if _, err := s.ValidateFinal(ctx, ports.ManagerValidateCommand{
			TenantID:    cmd.TenantID,
			TimesheetID: summary.ID,
			ManagerID:   cmd.ManagerID,
		}); err != nil {
			result.Failed = append(result.Failed, ports.ValidateAllFailure{
				TimesheetID: summary.ID,
				Reason:      err.Error(),
			})
			continue
		}
		result.Validated++
	}
	return result, nil
}
