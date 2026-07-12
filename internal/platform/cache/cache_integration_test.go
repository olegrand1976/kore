//go:build integration

package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

func TestInMemoryCacheGetSet(t *testing.T) {
	ctx := context.Background()
	c := cache.NewInMemoryCache()
	keys := cache.NewKeyBuilder("kore")
	tenant, _ := kernel.ParseTenantID("00000000-0000-4000-8000-000000000001")
	key := keys.Key(tenant, "test", "item")

	err := c.Set(ctx, key, map[string]string{"ok": "true"}, time.Minute)
	require.NoError(t, err)

	var out map[string]string
	found, err := c.Get(ctx, key, &out)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "true", out["ok"])
}
