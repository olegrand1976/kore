package app

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
)

var geminiModelPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

type platformService struct {
	repo               ports.PlatformRepository
	defaultGeminiModel string
	clock              func() time.Time
}

func NewPlatformService(repo ports.PlatformRepository, defaultGeminiModel string) ports.PlatformService {
	model := strings.TrimSpace(defaultGeminiModel)
	if model == "" {
		model = ports.DefaultGeminiModel
	}
	return &platformService{
		repo:               repo,
		defaultGeminiModel: model,
		clock:              time.Now,
	}
}

func (s *platformService) GetOverview(ctx context.Context) (ports.PlatformOverview, error) {
	tenants, err := s.repo.ListTenantsUsage(ctx)
	if err != nil {
		return ports.PlatformOverview{}, err
	}

	summary := ports.PlatformOverviewSummary{
		TenantsByStatus: make(map[string]int),
	}
	for _, t := range tenants {
		summary.TotalTenants++
		summary.TotalActiveUsers += t.ActiveUsers
		summary.TotalSeatLimit += t.SeatLimit
		if t.ActiveLast30d {
			summary.ActiveTenants30d++
		}
		summary.TenantsByStatus[t.SubscriptionStatus]++
	}

	return ports.PlatformOverview{
		Summary: summary,
		Tenants: tenants,
	}, nil
}

func (s *platformService) GetSettings(ctx context.Context) (ports.PlatformSettings, error) {
	settings, err := s.repo.GetPlatformSettings(ctx)
	if err != nil {
		return ports.PlatformSettings{
			GeminiModel: s.defaultGeminiModel,
			LLMProvider: "gemini",
		}, nil
	}
	if settings.GeminiModel == "" {
		settings.GeminiModel = s.defaultGeminiModel
	}
	settings.LLMProvider = "gemini"
	return settings, nil
}

func (s *platformService) UpdateSettings(ctx context.Context, cmd ports.UpdatePlatformSettingsCommand) (ports.PlatformSettings, error) {
	model, err := normalizeGeminiModel(cmd.GeminiModel)
	if err != nil {
		return ports.PlatformSettings{}, err
	}
	now := s.clock().UTC()
	if err := s.repo.SavePlatformSettings(ctx, model, cmd.ActorUserID, now); err != nil {
		return ports.PlatformSettings{}, err
	}
	return ports.PlatformSettings{
		GeminiModel: model,
		LLMProvider: "gemini",
		UpdatedAt:   &now,
	}, nil
}

func (s *platformService) CurrentGeminiModel(ctx context.Context) string {
	settings, err := s.repo.GetPlatformSettings(ctx)
	if err != nil || strings.TrimSpace(settings.GeminiModel) == "" {
		return s.defaultGeminiModel
	}
	return settings.GeminiModel
}

func normalizeGeminiModel(raw string) (string, error) {
	model := strings.TrimSpace(raw)
	if model == "" {
		return "", domain.ErrInvalidGeminiModel
	}
	if !geminiModelPattern.MatchString(model) {
		return "", domain.ErrInvalidGeminiModel
	}
	return model, nil
}
