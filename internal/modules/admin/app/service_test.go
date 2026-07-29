package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/admin/app"
	"github.com/kore/kore/internal/modules/admin/domain"
	"github.com/kore/kore/internal/modules/admin/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type fakeAdminRepo struct {
	params    map[string]domain.ParameterSet
	templates []domain.Template
	phones    []domain.PhoneDirectoryEntry
}

func newFakeAdminRepo() *fakeAdminRepo {
	return &fakeAdminRepo{params: map[string]domain.ParameterSet{}}
}

func (f *fakeAdminRepo) GetParameterSet(_ context.Context, tenant kernel.TenantID, code string) (domain.ParameterSet, error) {
	ps, ok := f.params[tenant.UUID().String()+":"+code]
	if !ok {
		return domain.ParameterSet{}, domain.ErrParameterSetNotFound
	}
	return ps, nil
}

func (f *fakeAdminRepo) SaveParameterSet(_ context.Context, ps domain.ParameterSet) error {
	f.params[ps.TenantID.UUID().String()+":"+ps.Code] = ps
	return nil
}

func (f *fakeAdminRepo) ListTemplates(_ context.Context, _ kernel.TenantID) ([]domain.Template, error) {
	return f.templates, nil
}

func (f *fakeAdminRepo) GetTemplate(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.Template, error) {
	for _, t := range f.templates {
		if t.ID == id {
			return t, nil
		}
	}
	return domain.Template{}, domain.ErrTemplateNotFound
}

func (f *fakeAdminRepo) SaveTemplate(_ context.Context, tmpl domain.Template) error {
	f.templates = append(f.templates, tmpl)
	return nil
}

func (f *fakeAdminRepo) ListPhoneDirectory(_ context.Context, _ kernel.TenantID) ([]domain.PhoneDirectoryEntry, error) {
	return f.phones, nil
}

func (f *fakeAdminRepo) SavePhoneEntry(_ context.Context, e domain.PhoneDirectoryEntry) error {
	f.phones = append(f.phones, e)
	return nil
}

func TestAdminService_UpsertParameters(t *testing.T) {
	repo := newFakeAdminRepo()
	svc := app.NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	ctx := context.Background()

	created, err := svc.UpsertParameters(ctx, ports.UpsertParametersCommand{
		TenantID: tenant,
		Code:     "security",
		Payload:  map[string]any{"mfa": true},
	})
	require.NoError(t, err)
	require.Equal(t, true, created.Payload["mfa"])

	updated, err := svc.UpsertParameters(ctx, ports.UpsertParametersCommand{
		TenantID: tenant,
		Code:     "security",
		Payload:  map[string]any{"mfa": false},
	})
	require.NoError(t, err)
	require.Equal(t, created.ID, updated.ID)
	require.Equal(t, false, updated.Payload["mfa"])
}

func TestAdminService_CreateTemplateAndPhone(t *testing.T) {
	repo := newFakeAdminRepo()
	svc := app.NewService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	ctx := context.Background()

	tmpl, err := svc.CreateTemplate(ctx, ports.CreateTemplateCommand{
		TenantID: tenant,
		Type:     "email",
		Name:     "welcome",
		Content:  map[string]any{"body": "hi"},
	})
	require.NoError(t, err)
	list, err := svc.ListTemplates(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, tmpl.ID, list[0].ID)

	phone, err := svc.CreatePhoneEntry(ctx, ports.CreatePhoneEntryCommand{
		TenantID:   tenant,
		Label:      "Hotline",
		Phone:      "+33999",
		Visibility: "public",
	})
	require.NoError(t, err)
	require.Equal(t, "public", phone.Visibility)
	phones, err := svc.ListPhoneDirectory(ctx, tenant)
	require.NoError(t, err)
	require.Len(t, phones, 1)
}
