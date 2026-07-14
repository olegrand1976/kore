package app

import (
	"context"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type noopRequestSettingsService struct{}

func (noopRequestSettingsService) Get(context.Context, kernel.TenantID) (domain.TenantRequestSettings, error) {
	return domain.TenantRequestSettings{
		ChannelsEnabled: domain.DefaultChannelsLegacy(),
		GuidesEnabled:   true,
	}, nil
}

func (noopRequestSettingsService) Update(context.Context, ports.UpdateRequestSettingsCommand) (domain.TenantRequestSettings, error) {
	return domain.TenantRequestSettings{}, nil
}

func (noopRequestSettingsService) IsChannelEnabled(context.Context, kernel.TenantID, kernel.RequestChannel) (bool, error) {
	return true, nil
}

func NoopRequestSettingsService() ports.RequestSettingsService {
	return noopRequestSettingsService{}
}

func NoopRequestChannelReader() kernel.RequestChannelReader {
	return noopRequestSettingsService{}
}
