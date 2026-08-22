package app

import (
	"fmt"
	"strings"
)

type TaigaConfig struct {
	BaseURL     string
	ProjectSlug string
}

func resolveTaigaExternalURL(data map[string]any, cfg TaigaConfig, ref *int) string {
	if permalink := stringFromAny(data["permalink"]); permalink != "" {
		return permalink
	}
	if u := stringFromAny(data["url"]); u != "" {
		return u
	}
	slug := taigaProjectSlug(data, cfg.ProjectSlug)
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" || slug == "" || ref == nil {
		return ""
	}
	return fmt.Sprintf("%s/project/%s/us/%d", base, slug, *ref)
}

func taigaProjectSlug(data map[string]any, fallback string) string {
	if raw, ok := data["project"]; ok {
		if m, ok := raw.(map[string]any); ok {
			if slug := stringFromAny(m["slug"]); slug != "" {
				return slug
			}
		}
	}
	return strings.TrimSpace(fallback)
}

func stringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "<nil>" {
			return ""
		}
		return s
	}
}
