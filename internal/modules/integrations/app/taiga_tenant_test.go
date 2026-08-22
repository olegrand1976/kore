package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveTaigaWebhookTenant_HeaderWins(t *testing.T) {
	mapping := TaigaKoreMapping{
		Projects: map[string]TaigaProjectMapping{
			"1": {KoreTenantID: "00000000-0000-4000-8000-0000000000a1"},
		},
	}
	got := ResolveTaigaWebhookTenant(
		"00000000-0000-4000-8000-000000000001",
		"default",
		mapping,
		map[string]any{"project": float64(1)},
	)
	require.Equal(t, "00000000-0000-4000-8000-000000000001", got)
}

func TestResolveTaigaWebhookTenant_FromMapping(t *testing.T) {
	mapping := TaigaKoreMapping{
		Projects: map[string]TaigaProjectMapping{
			"1": {KoreTenantID: "00000000-0000-4000-8000-000000000001"},
		},
	}
	got := ResolveTaigaWebhookTenant("", "00000000-0000-4000-8000-0000000000a1", mapping, map[string]any{
		"project": map[string]any{"id": float64(1)},
	})
	require.Equal(t, "00000000-0000-4000-8000-000000000001", got)
}

func TestResolveTaigaWebhookTenant_Default(t *testing.T) {
	got := ResolveTaigaWebhookTenant("", "00000000-0000-4000-8000-0000000000a1", TaigaKoreMapping{}, nil)
	require.Equal(t, "00000000-0000-4000-8000-0000000000a1", got)
}

func TestParseTaigaKoreMapping_Empty(t *testing.T) {
	m, err := ParseTaigaKoreMapping("")
	require.NoError(t, err)
	require.NotNil(t, m.Projects)
}
