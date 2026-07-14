package cra

import (
	"testing"
	"time"

	"github.com/google/uuid"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

func TestBuildGanttView_LabelsAndProgress(t *testing.T) {
	day1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	period, err := kernel.NewPeriod(day1, day2)
	if err != nil {
		t.Fatalf("period: %v", err)
	}
	view := BuildGanttView(period, []reportports.PlanningActivityRow{
		{
			UserID:       uuid.New(),
			Day:          day1,
			Minutes:      240,
			MissionID:    "mission-1",
			MissionLabel: "Projet Alpha",
			ClientLabel:  "ACME",
		},
		{
			UserID:       uuid.New(),
			Day:          day2,
			Minutes:      240,
			MissionID:    "mission-1",
			MissionLabel: "Projet Alpha",
			ClientLabel:  "ACME",
		},
	})
	if len(view.Items) != 1 {
		t.Fatalf("expected one gantt item, got %d", len(view.Items))
	}
	if view.Items[0].Label != "ACME · Projet Alpha" {
		t.Fatalf("unexpected label: %s", view.Items[0].Label)
	}
	if view.Items[0].Progress != 0.5 {
		t.Fatalf("expected progress 0.5, got %f", view.Items[0].Progress)
	}
}

func TestBuildPlanningView_GroupsByUser(t *testing.T) {
	userID := uuid.New()
	day := time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)
	period, _ := kernel.NewPeriod(day, day)
	view := BuildPlanningView(period, []reportports.PlanningActivityRow{
		{
			UserID:       userID,
			UserPrenom:   "Ada",
			UserNom:      "Lovelace",
			Day:          day,
			Minutes:      480,
			MissionLabel: "Support",
		},
	})
	if len(view.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(view.Rows))
	}
	if view.Rows[0].UserName != "Ada Lovelace" {
		t.Fatalf("unexpected user name: %s", view.Rows[0].UserName)
	}
	if len(view.Rows[0].Slots) != 1 || view.Rows[0].Slots[0].Hours != 8 {
		t.Fatalf("unexpected slot: %+v", view.Rows[0].Slots)
	}
}

func TestBuildPlanningView_LeaveSlot(t *testing.T) {
	userID := uuid.New()
	day := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	period, _ := kernel.NewPeriod(day, day)
	view := BuildPlanningView(period, []reportports.PlanningActivityRow{
		{
			UserID:       userID,
			Day:          day,
			Minutes:      0,
			MissionLabel: "Congé",
		},
	})
	if len(view.Rows) != 1 || len(view.Rows[0].Slots) != 1 {
		t.Fatalf("unexpected planning view: %+v", view)
	}
	if view.Rows[0].Slots[0].Label != "Congé" {
		t.Fatalf("expected leave label, got %s", view.Rows[0].Slots[0].Label)
	}
}

func TestGanttLabelFallback(t *testing.T) {
	if got := ganttLabel("", "", "abcdef12-3456"); got != "Mission abcdef12" {
		t.Fatalf("unexpected fallback label: %s", got)
	}
}
