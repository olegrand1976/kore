package app

import (
	"encoding/json"
	"fmt"
	"strings"
)

type TaigaProjectMapping struct {
	KoreTenantID string `json:"kore_tenant_id"`
}

type TaigaKoreMapping struct {
	Projects map[string]TaigaProjectMapping `json:"projects"`
}

func ParseTaigaKoreMapping(raw string) (TaigaKoreMapping, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return TaigaKoreMapping{Projects: map[string]TaigaProjectMapping{}}, nil
	}
	var mapping TaigaKoreMapping
	if err := json.Unmarshal([]byte(raw), &mapping); err != nil {
		return TaigaKoreMapping{}, fmt.Errorf("invalid taiga kore mapping: %w", err)
	}
	if mapping.Projects == nil {
		mapping.Projects = map[string]TaigaProjectMapping{}
	}
	return mapping, nil
}

func taigaProjectIDFromData(data map[string]any) string {
	if data == nil {
		return ""
	}
	raw, ok := data["project"]
	if !ok || raw == nil {
		return ""
	}
	switch t := raw.(type) {
	case float64:
		return fmt.Sprintf("%d", int(t))
	case int:
		return fmt.Sprintf("%d", t)
	case map[string]any:
		if id := intFromAny(t["id"]); id != nil {
			return fmt.Sprintf("%d", *id)
		}
	}
	return strings.TrimSpace(fmt.Sprint(raw))
}

func ResolveTaigaWebhookTenant(headerTenant, defaultTenant string, mapping TaigaKoreMapping, payloadData map[string]any) string {
	if tenant := strings.TrimSpace(headerTenant); tenant != "" {
		return tenant
	}
	projectID := taigaProjectIDFromData(payloadData)
	if projectID != "" {
		if entry, ok := mapping.Projects[projectID]; ok {
			if tenant := strings.TrimSpace(entry.KoreTenantID); tenant != "" {
				return tenant
			}
		}
	}
	return strings.TrimSpace(defaultTenant)
}

func ExtractTaigaWebhookData(body []byte) map[string]any {
	var payload struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	return payload.Data
}
