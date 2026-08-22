package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/integrations/domain"
	"github.com/kore/kore/internal/modules/integrations/ports"
	"github.com/kore/kore/pkg/kernel"
)

type TaigaService struct {
	repo ports.TaigaRepository
	cfg  TaigaConfig
}

func NewTaigaService(repo ports.TaigaRepository, cfg TaigaConfig) *TaigaService {
	return &TaigaService{repo: repo, cfg: cfg}
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
	if strings.TrimSpace(cmd.MatchMethod) == "" {
		return domain.ErrInvalidMatchMethod
	}
	return nil
}

type TaigaWebhookPayload struct {
	Action string         `json:"action"`
	Type   string         `json:"type"`
	Data   map[string]any `json:"data"`
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

func (s *TaigaService) FindByKoreDemand(ctx context.Context, tenant kernel.TenantID, demandID uuid.UUID) (domain.ExternalLink, error) {
	return s.repo.FindExternalLinkByKore(ctx, tenant, "demand", demandID)
}

func (s *TaigaService) HandleWebhook(ctx context.Context, tenant kernel.TenantID, body []byte) error {
	var payload TaigaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}
	data := payload.Data
	extRaw, ok := data["external_reference"]
	if !ok {
		return nil
	}
	extSlice, ok := extRaw.([]any)
	if !ok || len(extSlice) < 2 {
		return nil
	}
	if fmt.Sprint(extSlice[0]) != "kore" {
		return nil
	}
	koreID, err := uuid.Parse(fmt.Sprint(extSlice[1]))
	if err != nil {
		return nil
	}
	entityType := payload.Type
	if entityType == "" {
		entityType = "userstory"
	}
	entityID := fmt.Sprint(data["id"])
	projectID := intFromAny(data["project"])
	ref := intFromAny(data["ref"])
	externalURL := resolveTaigaExternalURL(data, s.cfg, ref)
	now := time.Now().UTC()
	link := domain.ExternalLink{
		TenantID:          tenant,
		Provider:          "taiga",
		ExternalType:      entityType,
		ExternalID:        entityID,
		ExternalProjectID: projectID,
		ExternalRef:       ref,
		ExternalURL:       externalURL,
		KoreEntityType:    "demand",
		KoreEntityID:      koreID,
		Metadata:          map[string]any{"action": payload.Action},
		LastSyncAt:        &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return s.repo.UpsertExternalLink(ctx, link)
}

func intFromAny(v any) *int {
	switch t := v.(type) {
	case float64:
		n := int(t)
		return &n
	case int:
		return &t
	default:
		return nil
	}
}
