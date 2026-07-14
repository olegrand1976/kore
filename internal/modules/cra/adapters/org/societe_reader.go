package org

import (
	"context"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/ports"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type SocieteReader struct {
	repo orgports.OrganizationRepository
}

func NewSocieteReader(repo orgports.OrganizationRepository) ports.SocieteCalendarReader {
	return &SocieteReader{repo: repo}
}

func (r *SocieteReader) SettingsForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (ports.SocieteCraSettings, error) {
	defaults := ports.SocieteCraSettings{
		WeekStartDay:       orgdomain.DefaultWeekStartDay,
		DayCapacityMinutes: orgdomain.DefaultDayCapacityMinutes,
		WeekSubmitPolicy:   orgdomain.DefaultWeekSubmitPolicy,
	}
	societeID, err := r.repo.ResolveSocieteIDForUser(ctx, tenant, userID)
	if err != nil {
		return defaults, nil
	}
	societe, err := r.repo.GetSociete(ctx, tenant, societeID)
	if err != nil {
		return defaults, nil
	}
	day := societe.WeekStartDay
	if day < 0 || day > 6 {
		day = orgdomain.DefaultWeekStartDay
	}
	cap := societe.DayCapacityMinutes
	if cap <= 0 || cap > 1440 {
		cap = orgdomain.DefaultDayCapacityMinutes
	}
	policy := societe.WeekSubmitPolicy
	if policy != "block" && policy != "warn" && policy != "none" {
		policy = orgdomain.DefaultWeekSubmitPolicy
	}
	return ports.SocieteCraSettings{
		WeekStartDay:       day,
		DayCapacityMinutes: cap,
		WeekSubmitPolicy:   policy,
		CraMailAuto:        societe.CraMailAuto,
		CraMailRecipients:  societe.CraMailRecipients,
		TaskTypesEnabled:   orgdomain.EffectiveTaskTypesEnabled(societe.TaskTypesEnabled),
	}, nil
}
