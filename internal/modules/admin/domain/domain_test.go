package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/admin/domain"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestNewParameterSet(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	ps := domain.NewParameterSet(tenant, "security", nil)
	require.Equal(t, "security", ps.Code)
	require.Equal(t, tenant, ps.TenantID)
	require.NotNil(t, ps.Payload)
	require.NotEqual(t, uuid.Nil, ps.ID)
}

func TestNewTemplate(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	tmpl := domain.NewTemplate(tenant, "email", "welcome", nil)
	require.True(t, tmpl.Active)
	require.Equal(t, "email", tmpl.Type)
	require.NotNil(t, tmpl.Content)
}

func TestNewPhoneDirectoryEntry(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	e := domain.NewPhoneDirectoryEntry(tenant, "Support", "+33123456789")
	require.Equal(t, "internal", e.Visibility)
	require.Equal(t, "Support", e.Label)
}
