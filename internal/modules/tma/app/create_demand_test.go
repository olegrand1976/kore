package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
	if repo.lastSaved.ID != got.ID {
		t.Fatalf("demand not saved")
	}
}
