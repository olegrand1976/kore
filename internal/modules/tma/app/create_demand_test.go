package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

func TestCreateDemandSucceedsWithoutDefaultBudget(t *testing.T) {
	repo := &fakeDemandRepo{}
	svc := NewService(repo, nil, nil)
	got, err := svc.CreateDemand(context.Background(), ports.CreateDemandCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		ApplicationID: uuid.New(),
		AuthorID:      uuid.New(),
		Subject:       "Suivi d'équipe",
		Description:   "",
		Priority:      "normal",
	})
	if err != nil {
		t.Fatalf("CreateDemand: %v", err)
	}
	if got.Subject != "Suivi d'équipe" {
		t.Fatalf("subject = %q", got.Subject)
	}
	if got.TicketNumber != 1 {
		t.Fatalf("ticketNumber = %d, want 1", got.TicketNumber)
	}
	if repo.lastSaved.ID != got.ID {
		t.Fatalf("demand not saved")
	}
}

func TestCreateDemandWithAssigneeSetsAssignedStatus(t *testing.T) {
	repo := &fakeDemandRepo{}
	svc := NewService(repo, nil, nil)
	assignee := uuid.New()
	got, err := svc.CreateDemand(context.Background(), ports.CreateDemandCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		ApplicationID: uuid.New(),
		AuthorID:      uuid.New(),
		AssigneeID:    &assignee,
		Subject:       "Affectée à la création",
		Priority:      "high",
	})
	if err != nil {
		t.Fatalf("CreateDemand: %v", err)
	}
	if got.AssigneeID == nil || *got.AssigneeID != assignee {
		t.Fatalf("assignee = %v, want %v", got.AssigneeID, assignee)
	}
	if got.Status != domain.DemandStatusAssigned {
		t.Fatalf("status = %q, want %s", got.Status, domain.DemandStatusAssigned)
	}
}

func TestCreateDemandWithAssigneeAndChefGatePreAssigns(t *testing.T) {
	repo := &fakeDemandRepo{}
	svc := NewService(repo, nil, nil)
	assignee := uuid.New()
	got, err := svc.CreateDemand(context.Background(), ports.CreateDemandCommand{
		TenantID:         kernel.NewTenantID(uuid.New()),
		ApplicationID:    uuid.New(),
		AuthorID:         uuid.New(),
		AssigneeID:       &assignee,
		Subject:          "Gate chef",
		Priority:         "normal",
		RequiresChefGate: true,
	})
	if err != nil {
		t.Fatalf("CreateDemand: %v", err)
	}
	if got.AssigneeID == nil || *got.AssigneeID != assignee {
		t.Fatalf("assignee = %v, want %v", got.AssigneeID, assignee)
	}
	if got.Status != domain.DemandStatusAwaitingCreation {
		t.Fatalf("status = %q, want %s", got.Status, domain.DemandStatusAwaitingCreation)
	}
	if got.Visible {
		t.Fatalf("expected not visible under chef gate")
	}
}
