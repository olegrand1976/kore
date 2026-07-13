package org

import (
	"context"

	"github.com/google/uuid"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

type SocieteReader struct {
	repo orgports.OrganizationRepository
}

func NewSocieteReader(repo orgports.OrganizationRepository) ports.SocieteCalendarReader {
	return &SocieteReader{repo: repo}
}

func (r *SocieteReader) WeekStartDayForUser(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (int, error) {
	societeID, err := r.repo.ResolveSocieteIDForUser(ctx, tenant, userID)
	if err != nil {
		return orgdomain.DefaultWeekStartDay, err
	}
	societe, err := r.repo.GetSociete(ctx, tenant, societeID)
	if err != nil {
		return orgdomain.DefaultWeekStartDay, err
	}
	day := societe.WeekStartDay
	if day < 0 || day > 6 {
		return orgdomain.DefaultWeekStartDay, nil
	}
	return day, nil
}
