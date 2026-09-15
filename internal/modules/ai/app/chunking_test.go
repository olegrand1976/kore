package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitTextChunks_empty(t *testing.T) {
	assert.Nil(t, SplitTextChunks(""))
	assert.Nil(t, SplitTextChunks("   \n\t"))
}

func TestSplitTextChunks_shortText(t *testing.T) {
	text := "Bonjour dossier TMA"
	got := SplitTextChunks(text)
	require.Len(t, got, 1)
	assert.Equal(t, text, got[0])
}

func TestSplitTextChunks_overlapWindows(t *testing.T) {
	runes := make([]rune, chunkTargetRunes+500)
	for i := range runes {
		runes[i] = 'a'
	}
	text := string(runes)
	chunks := SplitTextChunks(text)
	require.GreaterOrEqual(t, len(chunks), 2)
	for _, c := range chunks {
		assert.LessOrEqual(t, len([]rune(c)), chunkTargetRunes)
	}
	joined := strings.Join(chunks, "")
	assert.Greater(t, len([]rune(joined)), chunkTargetRunes)
}
