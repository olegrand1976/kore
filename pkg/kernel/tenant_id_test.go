package kernel

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTenantID_JSONRoundTrip(t *testing.T) {
	id := NewTenantID(uuid.MustParse("11111111-1111-1111-1111-111111111111"))

	raw, err := json.Marshal(id)
	require.NoError(t, err)
	require.Equal(t, `"11111111-1111-1111-1111-111111111111"`, string(raw))

	var decoded TenantID
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Equal(t, id, decoded)
}

func TestTenantID_MarshalJSON_NotEmptyObject(t *testing.T) {
	id := NewTenantID(uuid.New())
	raw, err := json.Marshal(map[string]any{"tenantId": id})
	require.NoError(t, err)

	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &payload))
	require.NotEqual(t, "{}", string(payload["tenantId"]))
}

func TestTenantID_UnmarshalJSON_Invalid(t *testing.T) {
	var id TenantID
	err := json.Unmarshal([]byte(`"not-a-uuid"`), &id)
	require.ErrorIs(t, err, ErrInvalidTenantID)

	err = json.Unmarshal([]byte(`{}`), &id)
	require.ErrorIs(t, err, ErrInvalidTenantID)
}
