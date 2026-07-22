package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
)

func TestNormalizeGeminiModel_valid(t *testing.T) {
	model, err := normalizeGeminiModel("gemini-3.6-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "gemini-3.6-flash" {
		t.Fatalf("got %q", model)
	}
}

func TestNormalizeGeminiModel_flashLite(t *testing.T) {
	model, err := normalizeGeminiModel("gemini-3.5-flash-lite")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "gemini-3.5-flash-lite" {
		t.Fatalf("got %q", model)
	}
}

func TestNormalizeGeminiModel_invalid(t *testing.T) {
	_, err := normalizeGeminiModel("../evil")
	if err == nil || err != domain.ErrInvalidGeminiModel {
		t.Fatalf("expected ErrInvalidGeminiModel, got %v", err)
	}
}

type stubPlatformRepo struct {
	settings ports.PlatformSettings
	err      error
}

func (r stubPlatformRepo) ListTenantsUsage(context.Context) ([]ports.TenantUsageSummary, error) {
	return nil, nil
}

func (r stubPlatformRepo) GetPlatformSettings(context.Context) (ports.PlatformSettings, error) {
	return r.settings, r.err
}

func (r stubPlatformRepo) SavePlatformSettings(context.Context, string, uuid.UUID, time.Time) error {
	return nil
}

func TestPlatformService_CurrentGeminiModel_fallback(t *testing.T) {
	s := &platformService{
		repo:               stubPlatformRepo{err: errors.New("db unavailable")},
		defaultGeminiModel: "gemini-3.6-flash",
	}
	if got := s.CurrentGeminiModel(context.Background()); got != "gemini-3.6-flash" {
		t.Fatalf("got %q", got)
	}
}
