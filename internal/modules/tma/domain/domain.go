package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

var (
	ErrDefaultBudgetRequired = errors.New("default budget required")
	ErrDemandNotVisible      = errors.New("demand not visible until chef utilisateur validation")
	ErrTransitionNotAllowed  = errors.New("transition not allowed")
	ErrDemandAlreadyResolved = errors.New("demand already resolved")
	ErrAnalysisNotFound      = errors.New("analysis not found")
)

type DemandType string

const DemandTypeIncident DemandType = "incident"

type DemandStatus string

const (
	DemandStatusAwaitingCreation DemandStatus = "en_attente_creation"
	DemandStatusOpen             DemandStatus = "ouverte"
	DemandStatusAssigned         DemandStatus = "affectee"
	DemandStatusInProgress       DemandStatus = "en_cours"
	DemandStatusResolved         DemandStatus = "resolue"
	DemandStatusRework           DemandStatus = "rework"
)

type Demand struct {
	ID                 uuid.UUID
	TenantID           kernel.TenantID
	ApplicationID      uuid.UUID
	Type               DemandType
	Subject            string
	WorkflowInstanceID *uuid.UUID
	AuthorID           uuid.UUID
	AssigneeID         *uuid.UUID
	Status             DemandStatus
	Visible            bool
	ConsumptionActive  bool
	RequiresChefGate   bool
	CreatedAt          time.Time
}

type AnalysisDossier struct {
	ID           uuid.UUID
	TenantID     kernel.TenantID
	DemandID     uuid.UUID
	Functional   string
	Technical    string
	Risks        string
	TestScenario string
}

type XmlExportRow struct {
	DemandID          uuid.UUID
	ApplicationID     uuid.UUID
	Type              string
	Subject           string
	Status            string
	AuthorID          uuid.UUID
	AssigneeID        *uuid.UUID
	CreatedAt         time.Time
	ResolvedAt        *time.Time
	EffortDays        float64
	EffortUO          float64
	Amount            int64
	ReleaseLabel      string
	DeliveryCode      string
	WorkflowState     string
	Visible           bool
	ConsumptionActive bool
	Comment           string
}

func NewDemand(tenant kernel.TenantID, appID, authorID uuid.UUID, subject string, requiresChefGate bool) Demand {
	status := DemandStatusOpen
	visible := true
	if requiresChefGate {
		status = DemandStatusAwaitingCreation
		visible = false
	}
	return Demand{
		ID:                uuid.New(),
		TenantID:          tenant,
		ApplicationID:     appID,
		Type:              DemandTypeIncident,
		Subject:           subject,
		AuthorID:          authorID,
		Status:            status,
		Visible:           visible,
		ConsumptionActive: visible,
		RequiresChefGate:  requiresChefGate,
		CreatedAt:         time.Now().UTC(),
	}
}

func (d *Demand) ValidateCreation() error {
	if !d.RequiresChefGate {
		return ErrTransitionNotAllowed
	}
	if d.Status != DemandStatusAwaitingCreation {
		return ErrTransitionNotAllowed
	}
	d.Status = DemandStatusOpen
	d.Visible = true
	d.ConsumptionActive = true
	return nil
}

func (d *Demand) Assign(assigneeID uuid.UUID) error {
	if !d.Visible {
		return ErrDemandNotVisible
	}
	d.AssigneeID = &assigneeID
	d.Status = DemandStatusAssigned
	return nil
}

func (d *Demand) TakeOver(userID uuid.UUID) error {
	if !d.Visible {
		return ErrDemandNotVisible
	}
	d.AssigneeID = &userID
	d.Status = DemandStatusInProgress
	return nil
}

func (d *Demand) Resolve() error {
	if !d.Visible {
		return ErrDemandNotVisible
	}
	if d.Status == DemandStatusResolved {
		return ErrDemandAlreadyResolved
	}
	d.Status = DemandStatusResolved
	return nil
}

func (d *Demand) Reopen(reason string) error {
	if !d.Visible {
		return ErrDemandNotVisible
	}
	d.Status = DemandStatusRework
	d.ConsumptionActive = true
	_ = reason
	return nil
}

func ToXmlExportRow(d Demand) XmlExportRow {
	return XmlExportRow{
		DemandID:          d.ID,
		ApplicationID:     d.ApplicationID,
		Type:              string(d.Type),
		Subject:           d.Subject,
		Status:            string(d.Status),
		AuthorID:          d.AuthorID,
		AssigneeID:        d.AssigneeID,
		CreatedAt:         d.CreatedAt,
		WorkflowState:     string(d.Status),
		Visible:           d.Visible,
		ConsumptionActive: d.ConsumptionActive,
	}
}
