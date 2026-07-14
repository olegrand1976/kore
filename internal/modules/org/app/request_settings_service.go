package app

import (
	"context"
	"time"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type requestSettingsService struct {
	repo ports.RequestSettingsRepository
}

func NewRequestSettingsService(repo ports.RequestSettingsRepository) *requestSettingsService {
	return &requestSettingsService{repo: repo}
}

func (s *requestSettingsService) Get(ctx context.Context, tenant kernel.TenantID) (domain.TenantRequestSettings, error) {
	settings, found, err := s.repo.GetTenantRequestSettings(ctx, tenant)
	if err != nil {
		return domain.TenantRequestSettings{}, err
	}
	if !found {
		return domain.DefaultTenantRequestSettings(tenant), nil
	}
	return settings, nil
}

func (s *requestSettingsService) Update(ctx context.Context, cmd ports.UpdateRequestSettingsCommand) (domain.TenantRequestSettings, error) {
	if !cmd.ChannelsEnabled.AtLeastOne() {
		return domain.TenantRequestSettings{}, domain.ErrInvalidRequestChannels
	}
	settings := domain.TenantRequestSettings{
		TenantID:        cmd.TenantID,
		ChannelsEnabled: cmd.ChannelsEnabled,
		GuidesEnabled:   cmd.GuidesEnabled,
		UpdatedAt:       time.Now().UTC(),
	}
	if err := s.repo.SaveTenantRequestSettings(ctx, settings); err != nil {
		return domain.TenantRequestSettings{}, err
	}
	return settings, nil
}

func (s *requestSettingsService) IsChannelEnabled(ctx context.Context, tenant kernel.TenantID, channel kernel.RequestChannel) (bool, error) {
	settings, err := s.Get(ctx, tenant)
	if err != nil {
		return false, err
	}
	return settings.ChannelsEnabled.IsEnabled(channel), nil
}

var (
	_ ports.RequestSettingsService = (*requestSettingsService)(nil)
	_ kernel.RequestChannelReader  = (*requestSettingsService)(nil)
)
