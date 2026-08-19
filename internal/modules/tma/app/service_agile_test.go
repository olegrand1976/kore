package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/tma/domain"
	"github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

type fakeAgileValidator struct {
	epicOK bool
}

func (f fakeAgileValidator) EpicBelongsToApplication(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID) (bool, error) {
	return f.epicOK, nil
}

func (f fakeAgileValidator) SprintBelongsToApplication(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func TestCreateDemandRejectsEpicFromOtherApplication(t *testing.T) {
	epicID := uuid.New()
	appID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	repo := &fakeDemandRepo{}
	svc := NewService(repo, nil, nil, WithAgileValidator(fakeAgileValidator{epicOK: false}))

	_, err := svc.CreateDemand(context.Background(), ports.CreateDemandCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		AuthorID:      uuid.New(),
		Subject:       "Story",
		EpicID:        &epicID,
	})
	if !errors.Is(err, domain.ErrEpicNotInApplication) {
		t.Fatalf("expected ErrEpicNotInApplication, got %v", err)
	}
}

func TestCreateDemandAcceptsEpicFromSameApplication(t *testing.T) {
	epicID := uuid.New()
	appID := uuid.New()
	tenant := kernel.NewTenantID(uuid.New())
	repo := &fakeDemandRepo{}
	svc := NewService(repo, nil, nil, WithAgileValidator(fakeAgileValidator{epicOK: true}))

	_, err := svc.CreateDemand(context.Background(), ports.CreateDemandCommand{
		TenantID:      tenant,
		ApplicationID: appID,
		AuthorID:      uuid.New(),
		Subject:       "Story",
		EpicID:        &epicID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
