package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/pkg/kernel"
)

type CreateMissionCommand struct {
	TenantID        kernel.TenantID
	ClientID        uuid.UUID
	StartDate       time.Time
	EndDate         *time.Time
	TJMAmount       int64
	Currency        string
	Technologies    []string
	ClientContact   string
	CollaboratorIDs []uuid.UUID
	CountryCode     string
}

type UpdateEndDateCommand struct {
	TenantID  kernel.TenantID
	MissionID uuid.UUID
	EndDate   time.Time
}

type MissionCollaborator struct {
	UserID uuid.UUID `json:"userId"`
	Login  string    `json:"login"`
	Prenom string    `json:"prenom"`
	Nom    string    `json:"nom"`
}

type MissionDetail struct {
	ID            uuid.UUID             `json:"id"`
	ClientID      uuid.UUID             `json:"clientId"`
	ClientName    string                `json:"clientName"`
	Status        string                `json:"status"`
	StartDate     time.Time             `json:"startDate"`
	EndDate       *time.Time            `json:"endDate,omitempty"`
	TJMAmount     int64                 `json:"tjmAmount"`
	Currency      string                `json:"currency"`
	Technologies  []string              `json:"technologies"`
	ClientContact string                `json:"clientContact"`
	CreatedAt     time.Time             `json:"createdAt"`
	Collaborators []MissionCollaborator `json:"collaborators"`
}

type MissionSummary struct {
	ID         uuid.UUID  `json:"id"`
	ClientID   uuid.UUID  `json:"clientId"`
	ClientName string     `json:"clientName"`
	Status     string     `json:"status"`
	StartDate  time.Time  `json:"startDate"`
	EndDate    *time.Time `json:"endDate,omitempty"`
	TJMAmount  int64      `json:"tjmAmount"`
	Currency   string     `json:"currency"`
}

type SSIIService interface {
	List(ctx context.Context, tenant kernel.TenantID) ([]domain.Mission, error)
	ListSummaries(ctx context.Context, tenant kernel.TenantID) ([]MissionSummary, error)
	Get(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error)
	GetDetail(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (MissionDetail, error)
	Create(ctx context.Context, cmd CreateMissionCommand) (domain.Mission, error)
	Stop(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error)
	UpdateEndDate(ctx context.Context, cmd UpdateEndDateCommand) (domain.Mission, error)
}

type SSIIRepository interface {
	SaveMission(ctx context.Context, m domain.Mission) error
	GetMission(ctx context.Context, tenant kernel.TenantID, id uuid.UUID) (domain.Mission, error)
	ListMissions(ctx context.Context, tenant kernel.TenantID) ([]domain.Mission, error)
	ListMissionSummaries(ctx context.Context, tenant kernel.TenantID) ([]MissionSummary, error)
	ListMissionCollaborators(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID) ([]MissionCollaborator, error)
	SaveMissionCollaborators(ctx context.Context, tenant kernel.TenantID, missionID uuid.UUID, userIDs []uuid.UUID) error
	GetClientName(ctx context.Context, tenant kernel.TenantID, clientID uuid.UUID) (string, error)
}
