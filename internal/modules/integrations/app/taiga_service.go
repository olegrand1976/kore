package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

type TaigaService struct {
	repo    ports.TaigaRepository
	cfg     TaigaConfig
	demands ports.TaigaDemandGate
	gateway ports.TaigaGateway
	apps    ports.ApplicationCreator
}

func NewTaigaService(repo ports.TaigaRepository, cfg TaigaConfig, demands ports.TaigaDemandGate, gateway ports.TaigaGateway, apps ports.ApplicationCreator) *TaigaService {
	return &TaigaService{repo: repo, cfg: cfg, demands: demands, gateway: gateway, apps: apps}
}

type UpsertUserMappingCommand struct {
	TenantID      kernel.TenantID
	TaigaUserID   int
	TaigaUsername string
	KoreUserID    uuid.UUID
	MatchMethod   string
}

func (cmd UpsertUserMappingCommand) Validate() error {
	if cmd.TaigaUserID <= 0 {
		return domain.ErrInvalidTaigaUserID
	}
	if cmd.KoreUserID == uuid.Nil {
		return domain.ErrInvalidKoreUserID
	}
	switch strings.TrimSpace(cmd.MatchMethod) {
	case domain.UserMatchMethodEmail, domain.UserMatchMethodManual:
		return nil
	default:
		return domain.ErrInvalidMatchMethod
	}
}

type TaigaWebhookPayload struct {
	Action     string         `json:"action"`
	Type       string         `json:"type"`
	Data       map[string]any `json:"data"`
	ValuesDiff map[string]any `json:"values_diff"`
	Change     map[string]any `json:"change"`
}

func (s *TaigaService) UpsertUserMapping(ctx context.Context, cmd UpsertUserMappingCommand) (domain.UserMapping, error) {
	if err := cmd.Validate(); err != nil {
		return domain.UserMapping{}, err
	}
	now := time.Now().UTC()
	mapping := domain.UserMapping{
		TenantID:         cmd.TenantID,
		Provider:         "taiga",
		ExternalUserID:   strconv.Itoa(cmd.TaigaUserID),
		ExternalUsername: cmd.TaigaUsername,
		KoreUserID:       cmd.KoreUserID,
		MatchMethod:      cmd.MatchMethod,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.repo.UpsertUserMapping(ctx, mapping); err != nil {
		return domain.UserMapping{}, err
	}
	return mapping, nil
}

func (s *TaigaService) ListUserMappings(ctx context.Context, tenant kernel.TenantID) ([]domain.UserMapping, error) {
	return s.repo.ListUserMappings(ctx, tenant, "taiga")
}

func (s *TaigaService) FindByKoreDemand(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.ExternalLink, error) {
	return s.repo.FindExternalLinkByKore(ctx, tenant, "demand", demandID)
}

func (s *TaigaService) HandleWebhook(ctx context.Context, tenant kernel.TenantID, body []byte) error {
	var payload TaigaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}
	data := payload.Data
	if data == nil {
		data = map[string]any{}
	}

	// Ignore change events that only stamp external_reference from Kore (anti-loop).
	if payload.Action == "change" && isExternalReferenceOnlyChange(payload, data) {
		return nil
	}

	entityType := payload.Type
	if entityType == "" {
		entityType = "userstory"
	}

	if extRaw, ok := data["external_reference"]; ok {
		if demandID, ok := parseKoreDemandRef(extRaw); ok {
			return s.linkExistingDemandFromWebhook(ctx, tenant, payload.Action, entityType, data, demandID)
		}
		if isKoreRefShape(extRaw) {
			return fmt.Errorf("%w: %q", domain.ErrInvalidKoreDemandID, fmt.Sprint(extRaw))
		}
	}

	// Phase 3: create Kore demand from new Taiga issue on a linked project.
	if entityType == "issue" && (payload.Action == "create" || payload.Action == "") {
		return s.createDemandFromWebhookIssue(ctx, tenant, data)
	}
	return nil
}

func (s *TaigaService) linkExistingDemandFromWebhook(
	ctx context.Context,
	tenant kernel.TenantID,
	action, entityType string,
	data map[string]any,
	koreID uuid.UUID,
) error {
	if s.demands != nil {
		exists, err := s.demands.KoreDemandExists(ctx, tenant, koreID)
		if err != nil {
			return err
		}
		if !exists {
			return domain.ErrKoreDemandNotFound
		}
	}
	entityID := fmt.Sprint(data["id"])
	projectID := intFromAny(data["project"])
	ref := intFromAny(data["ref"])
	externalURL := resolveTaigaExternalURL(data, s.cfg, ref)
	now := time.Now().UTC()
	extType := entityType
	if extType == "" {
		extType = "userstory"
	}
	link := domain.ExternalLink{
		TenantID:          tenant,
		Provider:          "taiga",
		ExternalType:      extType,
		ExternalID:        entityID,
		ExternalProjectID: projectID,
		ExternalRef:       ref,
		ExternalURL:       externalURL,
		KoreEntityType:    "demand",
		KoreEntityID:      koreID,
		Metadata:          map[string]any{"action": action, "syncOrigin": syncOriginTaiga},
		LastSyncAt:        &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return s.repo.UpsertExternalLink(ctx, link)
}

func (s *TaigaService) createDemandFromWebhookIssue(ctx context.Context, tenant kernel.TenantID, data map[string]any) error {
	projectID := intFromAny(data["project"])
	if projectID == nil || *projectID <= 0 {
		return nil
	}
	issueID := fmt.Sprint(data["id"])
	if issueID == "" || issueID == "<nil>" {
		return nil
	}
	if _, err := s.repo.FindExternalLinkByExternal(ctx, tenant, "taiga", externalTypeIssue, issueID); err == nil {
		return nil
	} else if err != domain.ErrExternalLinkNotFound {
		return err
	}
	if _, err := s.repo.FindApplicationByTaigaProjectID(ctx, tenant, strconv.Itoa(*projectID)); err != nil {
		if err == domain.ErrExternalLinkNotFound {
			return nil
		}
		return err
	}
	authorID, err := s.resolveSyncAuthor(ctx, tenant, nil)
	if err != nil {
		slog.Warn("taiga webhook skip create demand: no sync author",
			"tenantId", tenant.String(), "error", err)
		return nil
	}
	refPtr := intFromAny(data["ref"])
	ref := 0
	if refPtr != nil {
		ref = *refPtr
	}
	version := 1
	if v := intFromAny(data["version"]); v != nil {
		version = *v
	}
	issue := ports.TaigaIssue{
		ID:          atoiSafe(issueID),
		Ref:         ref,
		ProjectID:   *projectID,
		Subject:     strings.TrimSpace(fmt.Sprint(data["subject"])),
		Description: strings.TrimSpace(fmt.Sprint(data["description"])),
		Version:     version,
		Permalink:   strings.TrimSpace(fmt.Sprint(data["permalink"])),
	}
	if issue.ID <= 0 {
		return nil
	}
	_, err = s.PullIssueToKore(ctx, tenant, issue, authorID)
	return err
}

func isExternalReferenceOnlyChange(payload TaigaWebhookPayload, data map[string]any) bool {
	candidates := []map[string]any{nil, payload.ValuesDiff, payload.Change}
	if vals, ok := data["values_diff"].(map[string]any); ok {
		candidates[0] = vals
	}
	for _, vals := range candidates {
		if vals == nil {
			continue
		}
		if nested, ok := vals["values_diff"].(map[string]any); ok {
			vals = nested
		}
		if _, onlyRef := vals["external_reference"]; onlyRef && len(vals) == 1 {
			return true
		}
	}
	return false
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func intFromAny(v any) *int {
	switch t := v.(type) {
	case float64:
		n := int(t)
		return &n
	case int:
		return &t
	case json.Number:
		n64, err := t.Int64()
		if err != nil {
			return nil
		}
		n := int(n64)
		return &n
	default:
		return nil
	}
}
