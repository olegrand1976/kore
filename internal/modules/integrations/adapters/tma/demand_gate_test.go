package tma_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	integrationstma "github.com/kore/kore/internal/modules/integrations/adapters/tma"
	tmadomain "github.com/kore/kore/internal/modules/tma/domain"
	tmaports "github.com/kore/kore/internal/modules/tma/ports"
	"github.com/kore/kore/pkg/kernel"
)

type stubTMA struct {
	err             error
	lastCreate      tmaports.CreateDemandCommand
	createCalls     int
	createdDemandID uuid.UUID
}

func (s *stubTMA) CreateDemand(_ context.Context, cmd tmaports.CreateDemandCommand) (tmadomain.Demand, error) {
	s.createCalls++
	s.lastCreate = cmd
	id := s.createdDemandID
	if id == uuid.Nil {
		id = uuid.New()
	}
	return tmadomain.Demand{ID: id, TenantID: cmd.TenantID, ApplicationID: cmd.ApplicationID}, nil
}
func (s *stubTMA) Get(context.Context, kernel.TenantID, uuid.UUID) (tmadomain.Demand, error) {
	return tmadomain.Demand{}, s.err
}
func (s *stubTMA) ValidateCreation(context.Context, tmaports.ChefUtilisateurCommand) error {
	return nil
}
func (s *stubTMA) Assign(context.Context, tmaports.AssignCommand) error { return nil }
func (s *stubTMA) TakeOver(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID) error {
	return nil
}
func (s *stubTMA) AddAnalysis(context.Context, tmaports.AnalysisCommand) error { return nil }
func (s *stubTMA) Resolve(context.Context, kernel.TenantID, uuid.UUID, uuid.UUID) error {
	return nil
}
func (s *stubTMA) Reopen(context.Context, tmaports.ReworkCommand) error         { return nil }
func (s *stubTMA) SoftDelete(context.Context, kernel.TenantID, uuid.UUID) error { return nil }
func (s *stubTMA) List(context.Context, kernel.TenantID, tmaports.ExportFilter) ([]tmadomain.Demand, error) {
	return nil, nil
}
func (s *stubTMA) GetAnalysis(context.Context, kernel.TenantID, uuid.UUID) (tmadomain.AnalysisDossier, error) {
	return tmadomain.AnalysisDossier{}, nil
}
func (s *stubTMA) ExportXML(context.Context, tmaports.ExportFilter) ([]tmadomain.XmlExportRow, error) {
	return nil, nil
}

func TestDemandGate_Exists(t *testing.T) {
	gate := integrationstma.NewDemandGate(&stubTMA{err: nil})
	ok, err := gate.KoreDemandExists(context.Background(), kernel.NewTenantID(uuid.New()), uuid.New())
	if err != nil || !ok {
		t.Fatalf("exists: ok=%v err=%v", ok, err)
	}
}

func TestDemandGate_NotFound(t *testing.T) {
	gate := integrationstma.NewDemandGate(&stubTMA{err: tmadomain.ErrDemandNotFound})
	ok, err := gate.KoreDemandExists(context.Background(), kernel.NewTenantID(uuid.New()), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected false")
	}
}

func TestDemandGate_PropagatesError(t *testing.T) {
	gate := integrationstma.NewDemandGate(&stubTMA{err: errors.New("db down")})
	_, err := gate.KoreDemandExists(context.Background(), kernel.NewTenantID(uuid.New()), uuid.New())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDemandGate_CreateDemandFromTaigaSkipsOutboundSync(t *testing.T) {
	stub := &stubTMA{}
	gate := integrationstma.NewDemandGate(stub)
	_, err := gate.CreateDemandFromTaiga(
		context.Background(),
		kernel.NewTenantID(uuid.New()),
		uuid.New(),
		uuid.New(),
		"subj",
		"desc",
	)
	if err != nil {
		t.Fatal(err)
	}
	if stub.createCalls != 1 {
		t.Fatalf("createCalls=%d", stub.createCalls)
	}
	if !stub.lastCreate.SkipOutboundSync {
		t.Fatal("expected SkipOutboundSync=true for Taiga pull")
	}
}
