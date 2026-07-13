package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

type fakeDemandRepo struct {
	demand domain.Demand
}

func (r *fakeDemandRepo) Save(_ context.Context, _ domain.Demand) error { return nil }

func (r *fakeDemandRepo) Get(_ context.Context, _ kernel.TenantID, _ uuid.UUID) (domain.Demand, error) {
	return r.demand, nil
}

func (r *fakeDemandRepo) List(_ context.Context, _ kernel.TenantID, _ ports.ExportFilter) ([]domain.Demand, error) {
	return nil, nil
}

func (r *fakeDemandRepo) SaveAnalysis(_ context.Context, _ domain.AnalysisDossier) error { return nil }

func (r *fakeDemandRepo) GetAnalysis(_ context.Context, _ kernel.TenantID, _ uuid.UUID) (domain.AnalysisDossier, error) {
	return domain.AnalysisDossier{}, domain.ErrAnalysisNotFound
}

type captureNotifier struct {
	last ports.NotificationEvent
}

func (n *captureNotifier) Notify(_ context.Context, evt ports.NotificationEvent) error {
	n.last = evt
	return nil
}

func TestAddAnalysisPublishesNotification(t *testing.T) {
	ctx := context.Background()
	tenant := kernel.NewTenantID(uuid.New())
	demandID := uuid.New()
	actorID := uuid.New()

	repo := &fakeDemandRepo{
		demand: domain.Demand{
			ID:        demandID,
			TenantID:  tenant,
			Subject:   "Sujet TMA",
			Status:    domain.DemandStatusOpen,
			Visible:   true,
			AuthorID:  uuid.New(),
			Type:      domain.DemandTypeIncident,
			CreatedAt: time.Now().UTC(),
		},
	}
	notifier := &captureNotifier{}

	svc := NewService(repo, nil, nil, nil, WithNotifier(notifier))
	if err := svc.AddAnalysis(ctx, ports.AnalysisCommand{
		TenantID:     tenant,
		DemandID:     demandID,
		ActorID:      actorID,
		Functional:   "f",
		Technical:    "t",
		Risks:        "r",
		TestScenario: "ts",
	}); err != nil {
		t.Fatalf("AddAnalysis: %v", err)
	}

	if notifier.last.Trigger != "tma.analysis.updated" {
		t.Fatalf("expected trigger %q, got %q", "tma.analysis.updated", notifier.last.Trigger)
	}
	if notifier.last.TenantID != tenant {
		t.Fatalf("expected tenant to match")
	}
	if notifier.last.Vars["demandId"] != demandID.String() {
		t.Fatalf("expected demandId var")
	}
	if notifier.last.Vars["subject"] != "Sujet TMA" {
		t.Fatalf("expected subject var")
	}
	if notifier.last.Vars["authorId"] != actorID.String() {
		t.Fatalf("expected authorId var")
	}
}

