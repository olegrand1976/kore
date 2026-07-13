package app

import (
	"testing"

	"github.com/kore/kore/internal/platform/cache"
	"github.com/stretchr/testify/require"
)

func TestNewCRAService(t *testing.T) {
	svc := NewService(nil, cache.NewInMemoryCache(), cache.NewKeyBuilder("test"))
	require.NotNil(t, svc)
}
