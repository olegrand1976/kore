package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/cra/ports"
	"github.com/kore/kore/pkg/kernel"
)

type filterCaptureRepo struct {
	fakeCRARepo
	lastScope  *uuid.UUID
	lastFilter ports.TimesheetSummaryFilter
}

func (r *filterCaptureRepo) ListSummariesFiltered(_ context.Context, _ kernel.TenantID, scopeUserID *uuid.UUID, filter ports.TimesheetSummaryFilter) ([]domain.TimesheetSummary, error) {
	r.lastScope = scopeUserID
	r.lastFilter = filter
	return nil, nil
}

func TestListTimesheetSummaries_NonManagerScopedToViewer(t *testing.T) {
	repo := &filterCaptureRepo{}
	svc := &Service{repo: repo}
	tenant := kernel.NewTenantID(uuid.New())
	viewer := uuid.New()

	_, err := svc.ListTimesheetSummaries(context.Background(), tenant, viewer, false, ports.TimesheetSummaryFilter{
		Year:    "2026",
		MonthMM: "09",
		Limit:   24,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.lastScope == nil || *repo.lastScope != viewer {
		t.Fatalf("expected scope=%s, got %v", viewer, repo.lastScope)
	}
	if repo.lastFilter.Year != "2026" || repo.lastFilter.MonthMM != "09" {
		t.Fatalf("filter not forwarded: %+v", repo.lastFilter)
	}
}

func TestListTimesheetSummaries_ManagerPassesUserFilter(t *testing.T) {
	repo := &filterCaptureRepo{}
	svc := &Service{repo: repo}
	tenant := kernel.NewTenantID(uuid.New())
	viewer := uuid.New()
	target := uuid.New()

	_, err := svc.ListTimesheetSummaries(context.Background(), tenant, viewer, true, ports.TimesheetSummaryFilter{
		Year:    "2026",
		MonthMM: "08",
		UserID:  &target,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.lastScope == nil || *repo.lastScope != target {
		t.Fatalf("expected scope=%s (filter user), got %v", target, repo.lastScope)
	}
}

func TestListTimesheetSummaries_ManagerWithoutUserSeesTenant(t *testing.T) {
	repo := &filterCaptureRepo{}
	svc := &Service{repo: repo}
	tenant := kernel.NewTenantID(uuid.New())

	_, err := svc.ListTimesheetSummaries(context.Background(), tenant, uuid.New(), true, ports.TimesheetSummaryFilter{
		Year: "2026",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.lastScope != nil {
		t.Fatalf("expected nil scope for manager without userId, got %v", *repo.lastScope)
	}
}
