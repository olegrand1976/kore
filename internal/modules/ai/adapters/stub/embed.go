package stub

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"

	"github.com/kore/kore/internal/modules/ai/domain"
	"github.com/kore/kore/internal/modules/ai/ports"
)

// Embed returns deterministic unit-ish vectors for CI (dim 768).
func (p *Provider) Embed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, text := range texts {
		out[i] = DeterministicEmbedding(text)
	}
	return out, nil
}

func DeterministicEmbedding(text string) []float32 {
	sum := sha256.Sum256([]byte(text))
	vec := make([]float32, domain.EmbeddingDimensions)
	seed := binary.BigEndian.Uint64(sum[:8])
	var norm float64
	for i := range vec {
		// xorshift-ish from seed + index
		seed = seed*6364136223846793005 + uint64(i+1)
		v := float32((seed%20001)-10000) / 10000
		vec[i] = v
		norm += float64(v * v)
	}
	if norm > 0 {
		inv := float32(1 / math.Sqrt(norm))
		for i := range vec {
			vec[i] *= inv
		}
	}
	return vec
}

var _ ports.EmbeddingsProvider = (*Provider)(nil)
