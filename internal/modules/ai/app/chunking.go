package app

import (
	"strings"
)

const (
	chunkTargetRunes = 1800 // ~600–700 tokens FR
	chunkOverlapRunes = 200
	maxChunksPerFile = 40
)

// SplitTextChunks splits text into overlapping rune windows for embedding.
func SplitTextChunks(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	runes := []rune(text)
	if len(runes) <= chunkTargetRunes {
		return []string{string(runes)}
	}
	var out []string
	step := chunkTargetRunes - chunkOverlapRunes
	if step < 1 {
		step = chunkTargetRunes
	}
	for start := 0; start < len(runes) && len(out) < maxChunksPerFile; start += step {
		end := start + chunkTargetRunes
		if end > len(runes) {
			end = len(runes)
		}
		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			out = append(out, chunk)
		}
		if end >= len(runes) {
			break
		}
	}
	return out
}
