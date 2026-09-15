package stub

import (
	"context"
	"testing"

	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeterministicEmbedding_unitNorm(t *testing.T) {
	vec := DeterministicEmbedding("hello")
	require.Len(t, vec, domain.EmbeddingDimensions)
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}
	assert.InDelta(t, 1.0, sum, 1e-4)
}

func TestProviderEmbed_deterministic(t *testing.T) {
	p := NewProvider()
	out, err := p.Embed(context.Background(), []string{"a", "b", "a"})
	require.NoError(t, err)
	require.Len(t, out, 3)
	assert.Equal(t, out[0], out[2])
	assert.NotEqual(t, out[0], out[1])
}
