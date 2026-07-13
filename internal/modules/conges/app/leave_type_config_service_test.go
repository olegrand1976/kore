package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	congesapp "github.com/kore/kore/internal/modules/conges/app"
	"github.com/kore/kore/internal/modules/conges/domain"
	"github.com/kore/kore/internal/modules/conges/ports"
	orgdomain "github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLeaveTypeRepo struct {
	byCode map[string]domain.LeaveTypeConfig
	items  []domain.LeaveTypeConfig
}

func (m *mockLeaveTypeRepo) ListBySociete(_ context.Context, _ kernel.TenantID, _ uuid.UUID, activeOnly bool) ([]domain.LeaveTypeConfig, error) {
	if activeOnly {
		var out []domain.LeaveTypeConfig
		for _, item := range m.items {
			if item.Active {
				out = append(out, item)
			}
		}
		return out, nil
	}
	return m.items, nil
}

func (m *mockLeaveTypeRepo) Get(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.LeaveTypeConfig, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.LeaveTypeConfig{}, domain.ErrLeaveTypeNotFound
}

func (m *mockLeaveTypeRepo) GetByCode(_ context.Context, _ kernel.TenantID, _ uuid.UUID, code string) (domain.LeaveTypeConfig, error) {
	if cfg, ok := m.byCode[code]; ok {
		return cfg, nil
	}
	return domain.LeaveTypeConfig{}, domain.ErrLeaveTypeNotFound
}

func (m *mockLeaveTypeRepo) Save(_ context.Context, cfg domain.LeaveTypeConfig) error {
	m.byCode[cfg.Code] = cfg
	found := false
	for i, item := range m.items {
		if item.ID == cfg.ID || item.Code == cfg.Code {
			m.items[i] = cfg
			found = true
			break
		}
	}
	if !found {
		m.items = append(m.items, cfg)
	}
	return nil
}

func (m *mockLeaveTypeRepo) Delete(_ context.Context, _ kernel.TenantID, id uuid.UUID) error {
	for i, item := range m.items {
		if item.ID == id {
			delete(m.byCode, item.Code)
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return domain.ErrLeaveTypeNotFound
}

func (m *mockLeaveTypeRepo) IsCodeUsed(_ context.Context, _ kernel.TenantID, _ uuid.UUID, code string) (bool, error) {
	return code == "conges_payes", nil
}

func (m *mockLeaveTypeRepo) UpsertDefaults(_ context.Context, _ kernel.TenantID, societeID uuid.UUID, templates []domain.LeaveTypeTemplate) error {
	for _, tpl := range templates {
		cfg := domain.LeaveTypeConfig{
			ID:            uuid.New(),
			SocieteID:     societeID,
			Code:          tpl.Code,
			Label:         tpl.Label,
			TracksBalance: tpl.TracksBalance,
			Active:        true,
			SortOrder:     tpl.SortOrder,
		}
		if err := m.Save(context.Background(), cfg); err != nil {
			return err
		}
	}
	return nil
}

type mockOrgReader struct {
	societe   orgdomain.Societe
	societeID uuid.UUID
}

func (m *mockOrgReader) GetSociete(_ context.Context, _ kernel.TenantID, _ uuid.UUID) (orgdomain.Societe, error) {
	return m.societe, nil
}

func (m *mockOrgReader) ResolveSocieteIDForUser(_ context.Context, _ kernel.TenantID, _ uuid.UUID) (uuid.UUID, error) {
	return m.societeID, nil
}

func TestLeaveTypeConfigService_ValidateTypeForUser(t *testing.T) {
	societeID := uuid.New()
	repo := &mockLeaveTypeRepo{
		byCode: map[string]domain.LeaveTypeConfig{
			"rtt": {Code: "rtt", Active: true, SocieteID: societeID},
		},
	}
	svc := congesapp.NewLeaveTypeConfigService(repo, &mockOrgReader{societeID: societeID})
	tenant := kernel.NewTenantID(uuid.New())
	userID := uuid.New()

	require.NoError(t, svc.ValidateTypeForUser(context.Background(), tenant, userID, "rtt"))
	assert.ErrorIs(t, svc.ValidateTypeForUser(context.Background(), tenant, userID, "unknown"), domain.ErrUnknownLeaveType)
}

func TestLeaveTypeConfigService_ResetDefaults_FR(t *testing.T) {
	societeID := uuid.New()
	repo := &mockLeaveTypeRepo{byCode: map[string]domain.LeaveTypeConfig{}}
	svc := congesapp.NewLeaveTypeConfigService(repo, &mockOrgReader{
		societeID: societeID,
		societe:   orgdomain.Societe{ID: societeID, Pays: "FR"},
	})
	tenant := kernel.NewTenantID(uuid.New())

	items, err := svc.ResetDefaults(context.Background(), ports.ResetLeaveTypeConfigsCommand{
		TenantID:  tenant,
		SocieteID: societeID,
	})
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, "conges_payes", items[0].Code)
}

func TestLeaveTypeConfigService_DeleteInUseDeactivates(t *testing.T) {
	societeID := uuid.New()
	id := uuid.New()
	repo := &mockLeaveTypeRepo{
		byCode: map[string]domain.LeaveTypeConfig{
			"conges_payes": {ID: id, Code: "conges_payes", Active: true, SocieteID: societeID},
		},
		items: []domain.LeaveTypeConfig{
			{ID: id, Code: "conges_payes", Active: true, SocieteID: societeID},
		},
	}
	svc := congesapp.NewLeaveTypeConfigService(repo, &mockOrgReader{societeID: societeID})
	tenant := kernel.NewTenantID(uuid.New())

	err := svc.Delete(context.Background(), tenant, id)
	assert.ErrorIs(t, err, domain.ErrLeaveTypeInUse)
	assert.False(t, repo.byCode["conges_payes"].Active)
}
