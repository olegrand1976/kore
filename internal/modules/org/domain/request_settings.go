package domain

import (
	"errors"
	"time"

	"github.com/kore/kore/pkg/kernel"
)

var ErrInvalidRequestChannels = errors.New("at least one request channel must be enabled")

type ChannelsEnabled struct {
	TMA         bool `json:"tma"`
	Support     bool `json:"support"`
	Maintenance bool `json:"maintenance"`
}

func DefaultChannelsForNewTenant() ChannelsEnabled {
	return ChannelsEnabled{TMA: true, Support: false, Maintenance: false}
}

func DefaultChannelsLegacy() ChannelsEnabled {
	return ChannelsEnabled{TMA: true, Support: true, Maintenance: true}
}

func (c ChannelsEnabled) AtLeastOne() bool {
	return c.TMA || c.Support || c.Maintenance
}

func (c ChannelsEnabled) IsEnabled(channel kernel.RequestChannel) bool {
	switch channel {
	case kernel.RequestChannelTMA:
		return c.TMA
	case kernel.RequestChannelSupport:
		return c.Support
	case kernel.RequestChannelMaintenance:
		return c.Maintenance
	default:
		return false
	}
}

type TenantRequestSettings struct {
	TenantID        kernel.TenantID `json:"tenantId"`
	ChannelsEnabled ChannelsEnabled `json:"channelsEnabled"`
	GuidesEnabled   bool            `json:"guidesEnabled"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

func DefaultTenantRequestSettings(tenant kernel.TenantID) TenantRequestSettings {
	return TenantRequestSettings{
		TenantID:        tenant,
		ChannelsEnabled: DefaultChannelsForNewTenant(),
		GuidesEnabled:   true,
		UpdatedAt:       time.Now().UTC(),
	}
}
