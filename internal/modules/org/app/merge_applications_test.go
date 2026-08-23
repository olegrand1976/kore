package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
)

type taigaLinkReaderStub struct {
	linked map[uuid.UUID]bool
}

func (s *taigaLinkReaderStub) IsApplicationTaigaLinked(_ context.Context, _ kernel.TenantID, applicationID uuid.UUID) (bool, error) {
	return s.linked[applicationID], nil
}

type mergeRepoStub struct {
	refreshUserRepo
	mergedAbsorbed  uuid.UUID
	mergedReference uuid.UUID
	mergeErr        error
	apps            map[uuid.UUID]domain.Application
	artifactCounts  map[uuid.UUID]int
	auditAction     string
	auditPayload    map[string]interface{}
}

func (r *mergeRepoStub) GetApplication(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.Application, error) {
	if app, ok := r.apps[id]; ok {
		return app, nil
	}
	return domain.Application{ID: id, Active: true, MethodologyProfile: domain.MethodologyPSA}, nil
}

func (r *mergeRepoStub) CountProjectArtifacts(_ context.Context, _ kernel.TenantID, applicationID uuid.UUID) (int, error) {
	if r.artifactCounts == nil {
		return 0, nil
	}
	return r.artifactCounts[applicationID], nil
}

func (r *mergeRepoStub) MergeApplications(_ context.Context, _ kernel.TenantID, absorbedID, referenceID uuid.UUID) (domain.Application, error) {
	if r.mergeErr != nil {
		return domain.Application{}, r.mergeErr
	}
	r.mergedAbsorbed = absorbedID
	r.mergedReference = referenceID
	return domain.Application{ID: referenceID}, nil
}

func (r *mergeRepoStub) RecordAdminAuditEvent(_ context.Context, _ kernel.TenantID, _ uuid.UUID, action string, payload map[string]interface{}) error {
	r.auditAction = action
	r.auditPayload = payload
	return nil
}

func TestMergeApplications_rejectsBothTaigaLinked(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	taigaID := uuid.New()
	otherTaigaID := uuid.New()

	svc := NewOrganizationService(&mergeRepoStub{}, &taigaLinkReaderStub{
		linked: map[uuid.UUID]bool{taigaID: true, otherTaigaID: true},
	})

	_, err := svc.MergeApplications(context.Background(), ports.MergeApplicationsCommand{
		TenantID:            tenant,
		SourceApplicationID: taigaID,
		TargetApplicationID: otherTaigaID,
	})
	if !errors.Is(err, domain.ErrApplicationsMergeBothTaigaLinked) {
		t.Fatalf("err = %v, want %v", err, domain.ErrApplicationsMergeBothTaigaLinked)
	}
}

func TestMergeApplications_swapsTaigaAsReference(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	taigaID := uuid.New()
	manualID := uuid.New()
	repo := &mergeRepoStub{}

	svc := NewOrganizationService(repo, &taigaLinkReaderStub{
		linked: map[uuid.UUID]bool{taigaID: true},
	})

	_, err := svc.MergeApplications(context.Background(), ports.MergeApplicationsCommand{
		TenantID:            tenant,
		SourceApplicationID: taigaID,
		TargetApplicationID: manualID,
	})
	if err != nil {
		t.Fatalf("MergeApplications: %v", err)
	}
	if repo.mergedReference != taigaID {
		t.Fatalf("reference = %v, want %v", repo.mergedReference, taigaID)
	}
	if repo.mergedAbsorbed != manualID {
		t.Fatalf("absorbed = %v, want %v", repo.mergedAbsorbed, manualID)
	}
}

func TestMergeApplications_rejectsMethodologyConflictWithArtifacts(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	psaID := uuid.New()
	scrumID := uuid.New()
	repo := &mergeRepoStub{
		apps: map[uuid.UUID]domain.Application{
			psaID:   {ID: psaID, Active: true, MethodologyProfile: domain.MethodologyPSA},
			scrumID: {ID: scrumID, Active: true, MethodologyProfile: domain.MethodologyAgileScrum},
		},
		artifactCounts: map[uuid.UUID]int{scrumID: 1},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.MergeApplications(context.Background(), ports.MergeApplicationsCommand{
		TenantID:            tenant,
		SourceApplicationID: psaID,
		TargetApplicationID: scrumID,
	})
	if !errors.Is(err, domain.ErrApplicationsMergeMethodologyConflict) {
		t.Fatalf("err = %v, want %v", err, domain.ErrApplicationsMergeMethodologyConflict)
	}
}

func TestMergeApplications_rejectsInactiveApplication(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	absorbedID := uuid.New()
	referenceID := uuid.New()
	repo := &mergeRepoStub{
		apps: map[uuid.UUID]domain.Application{
			absorbedID:  {ID: absorbedID, Active: false, Libelle: "Inactive"},
			referenceID: {ID: referenceID, Active: true, Libelle: "Reference"},
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.MergeApplications(context.Background(), ports.MergeApplicationsCommand{
		TenantID:            tenant,
		SourceApplicationID: absorbedID,
		TargetApplicationID: referenceID,
	})
	if !errors.Is(err, domain.ErrApplicationsMergeInactiveApplication) {
		t.Fatalf("err = %v, want %v", err, domain.ErrApplicationsMergeInactiveApplication)
	}
}

func TestMergeApplications_recordsAuditEvent(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	actorID := uuid.New()
	absorbedID := uuid.New()
	referenceID := uuid.New()
	repo := &mergeRepoStub{
		apps: map[uuid.UUID]domain.Application{
			absorbedID:  {ID: absorbedID, Active: true, Libelle: "Absorbed"},
			referenceID: {ID: referenceID, Active: true, Libelle: "Reference"},
		},
	}
	svc := NewOrganizationService(repo, nil)

	_, err := svc.MergeApplications(context.Background(), ports.MergeApplicationsCommand{
		TenantID:            tenant,
		ActorUserID:         actorID,
		SourceApplicationID: absorbedID,
		TargetApplicationID: referenceID,
	})
	if err != nil {
		t.Fatalf("MergeApplications: %v", err)
	}
	if repo.auditAction != "applications.merge" {
		t.Fatalf("audit action = %q", repo.auditAction)
	}
	if repo.auditPayload["absorbedApplicationId"] != absorbedID.String() {
		t.Fatalf("audit payload absorbed = %v", repo.auditPayload)
	}
}
