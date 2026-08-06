package app

import (
	"context"
	"testing"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type requestSettingsRepoStub struct {
	settings domain.TenantRequestSettings
	found    bool
}

func (s *requestSettingsRepoStub) GetTenantRequestSettings(_ context.Context, tenant kernel.TenantID) (domain.TenantRequestSettings, bool, error) {
	if !s.found {
		return domain.TenantRequestSettings{}, false, nil
	}
	out := s.settings
	out.TenantID = tenant
	return out, true, nil
}

func (s *requestSettingsRepoStub) SaveTenantRequestSettings(_ context.Context, settings domain.TenantRequestSettings) error {
	s.settings = settings
	s.found = true
	return nil
}

func TestRequestSettingsUpdateRequiresOneChannel(t *testing.T) {
	repo := &requestSettingsRepoStub{}
	svc := NewRequestSettingsService(repo)
	tenant := kernel.TenantID{}
	_, err := svc.Update(context.Background(), ports.UpdateRequestSettingsCommand{
		TenantID:        tenant,
		ChannelsEnabled: domain.ChannelsEnabled{},
		GuidesEnabled:   true,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRequestSettingsIsChannelEnabled(t *testing.T) {
	repo := &requestSettingsRepoStub{
		found: true,
		settings: domain.TenantRequestSettings{
			ChannelsEnabled: domain.ChannelsEnabled{TMA: true, Support: false, Maintenance: true},
		},
	}
	svc := NewRequestSettingsService(repo)
	tenant := kernel.TenantID{}
	ok, err := svc.IsChannelEnabled(context.Background(), tenant, kernel.RequestChannelSupport)
	if err != nil || ok {
		t.Fatalf("support should be disabled, got ok=%v err=%v", ok, err)
	}
	ok, err = svc.IsChannelEnabled(context.Background(), tenant, kernel.RequestChannelMaintenance)
	if err != nil || !ok {
		t.Fatalf("maintenance should be enabled, got ok=%v err=%v", ok, err)
	}
}

func TestRequestSettingsInvoicingToggle(t *testing.T) {
	repo := &requestSettingsRepoStub{}
	svc := NewRequestSettingsService(repo)
	tenant := kernel.TenantID{}

	ok, err := svc.IsInvoicingEnabled(context.Background(), tenant)
	if err != nil || ok {
		t.Fatalf("default invoicing should be disabled, got ok=%v err=%v", ok, err)
	}

	enabled := true
	updated, err := svc.Update(context.Background(), ports.UpdateRequestSettingsCommand{
		TenantID:         tenant,
		ChannelsEnabled:  domain.ChannelsEnabled{TMA: true},
		GuidesEnabled:    true,
		InvoicingEnabled: &enabled,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !updated.InvoicingEnabled {
		t.Fatal("expected invoicing enabled on update result")
	}
	ok, err = svc.IsInvoicingEnabled(context.Background(), tenant)
	if err != nil || !ok {
		t.Fatalf("invoicing should be enabled, got ok=%v err=%v", ok, err)
	}

	// Omitting InvoicingEnabled must preserve the current value.
	updated, err = svc.Update(context.Background(), ports.UpdateRequestSettingsCommand{
		TenantID:        tenant,
		ChannelsEnabled: domain.ChannelsEnabled{TMA: true, Support: true},
		GuidesEnabled:   false,
	})
	if err != nil {
		t.Fatalf("update without invoicing flag: %v", err)
	}
	if !updated.InvoicingEnabled {
		t.Fatal("expected invoicing to stay enabled when field omitted")
	}
	if updated.GuidesEnabled {
		t.Fatal("expected guides disabled")
	}
}
