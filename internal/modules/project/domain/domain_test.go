package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/project/domain"
	"github.com/kore/kore/pkg/kernel"
)

func TestSprintStartClose(t *testing.T) {
	t.Parallel()
	s := domain.NewSprint(kernel.NewTenantID(uuid.New()), uuid.New(), "S1", "Goal",
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC), nil)
	if err := s.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := s.Close(time.Now().UTC()); err != nil {
		t.Fatalf("close: %v", err)
	}
	if s.Status != domain.SprintStatusClosed {
		t.Fatalf("status %s", s.Status)
	}
}

func TestValidateKanbanConfig(t *testing.T) {
	t.Parallel()
	valid := []domain.KanbanColumn{
		{StateCode: "ouverte", Label: "Open"},
		{StateCode: "en_cours", Label: "Doing", WipLimit: intPtr(3)},
	}
	if err := domain.ValidateKanbanConfig(valid); err != nil {
		t.Fatalf("valid config: %v", err)
	}
	dup := []domain.KanbanColumn{
		{StateCode: "ouverte"},
		{StateCode: "ouverte"},
	}
	if err := domain.ValidateKanbanConfig(dup); err != domain.ErrInvalidKanbanConfig {
		t.Fatalf("expected invalid kanban config, got %v", err)
	}
}

func intPtr(n int) *int { return &n }

func TestValidateStoryPoints(t *testing.T) {
	t.Parallel()
	v := int16(5)
	if err := domain.ValidateStoryPoints(&v); err != nil {
		t.Fatalf("valid points: %v", err)
	}
	bad := int16(-1)
	if err := domain.ValidateStoryPoints(&bad); err == nil {
		t.Fatal("expected invalid story points")
	}
}
