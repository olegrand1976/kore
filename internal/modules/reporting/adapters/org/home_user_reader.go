package org

import (
	"context"

	"github.com/google/uuid"
	orgports "github.com/kore/kore/internal/modules/org/ports"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type HomeUserReader struct {
	users orgports.UserService
}

func NewHomeUserReader(users orgports.UserService) reportports.HomeUserReader {
	return &HomeUserReader{users: users}
}

func (r *HomeUserReader) GetCraRequis(ctx context.Context, tenant kernel.TenantID, userID uuid.UUID) (bool, error) {
	detail, err := r.users.GetUser(ctx, tenant, userID)
	if err != nil {
		return false, err
	}
	return detail.CraRequis, nil
}

var _ reportports.HomeUserReader = (*HomeUserReader)(nil)
