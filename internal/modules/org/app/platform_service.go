package app

import (
	"context"

	"github.com/kore/kore/internal/modules/org/ports"
)

type platformService struct {
	repo ports.PlatformRepository
}

func NewPlatformService(repo ports.PlatformRepository) ports.PlatformService {
	return &platformService{repo: repo}
}

func (s *platformService) GetOverview(ctx context.Context) (ports.PlatformOverview, error) {
	tenants, err := s.repo.ListTenantsUsage(ctx)
	if err != nil {
		return ports.PlatformOverview{}, err
	}

	summary := ports.PlatformOverviewSummary{
		TenantsByStatus: make(map[string]int),
	}
	for _, t := range tenants {
		summary.TotalTenants++
		summary.TotalActiveUsers += t.ActiveUsers
		summary.TotalSeatLimit += t.SeatLimit
		if t.ActiveLast30d {
			summary.ActiveTenants30d++
		}
		summary.TenantsByStatus[t.SubscriptionStatus]++
	}

	return ports.PlatformOverview{
		Summary: summary,
		Tenants: tenants,
	}, nil
}
