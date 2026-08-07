package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	cradomain "github.com/kore/kore/internal/modules/cra/domain"
	"github.com/kore/kore/internal/modules/ssii/domain"
	"github.com/kore/kore/internal/modules/ssii/ports"
	"github.com/kore/kore/pkg/kernel"
)

type fakeFeeder struct {
	lines []ports.ProposedMissionLine
}

func (f *fakeFeeder) ProposeLines(_ context.Context, lines []ports.ProposedMissionLine) error {
	f.lines = append(f.lines, lines...)
	return nil
}

type fakeCalendar struct {
	blocked map[string]bool
}

func (f *fakeCalendar) IsHolidayOrLeave(_ context.Context, _ kernel.TenantID, _ uuid.UUID, day time.Time, _ string) (bool, error) {
	return f.blocked[day.Format("2006-01-02")], nil
}

func TestPrefillMissionDays_SkipsBlockedDays(t *testing.T) {
	feeder := &fakeFeeder{}
	calendar := &fakeCalendar{blocked: map[string]bool{
		"2026-07-14": true,
	}}
	svc := &service{feeder: feeder, calendar: calendar}
	mission := domain.Mission{
		ID:        uuid.New(),
		TenantID:  kernel.NewTenantID(uuid.New()),
		StartDate: time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC),
	}
	userID := uuid.New()
	err := svc.prefillMissionDays(context.Background(), mission, []uuid.UUID{userID}, "FR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(feeder.lines) == 0 {
		t.Fatal("expected prefill lines")
	}
	for _, line := range feeder.lines {
		if line.Day.Format("2006-01-02") == "2026-07-14" {
			t.Fatal("blocked holiday should not be prefilled")
		}
		_ = cradomain.Month(line.Month)
	}
}

func TestCreate_rejectsInvalidRateUnit(t *testing.T) {
	svc := NewService(&noopRepo{}, nil, nil, nil)
	_, err := svc.Create(context.Background(), ports.CreateMissionCommand{
		TenantID:        kernel.NewTenantID(uuid.New()),
		ClientID:        uuid.New(),
		StartDate:       time.Now().UTC(),
		RateUnit:        "weekly",
		CollaboratorIDs: []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, domain.ErrInvalidRateUnit) {
		t.Fatalf("err = %v, want ErrInvalidRateUnit", err)
	}
}

func TestCreate_rejectsUnknownClientContact(t *testing.T) {
	svc := NewService(&noopRepo{}, nil, nil, nil)
	_, err := svc.Create(context.Background(), ports.CreateMissionCommand{
		TenantID:         kernel.NewTenantID(uuid.New()),
		ClientID:         uuid.New(),
		StartDate:        time.Now().UTC(),
		RateUnit:         "tjm",
		CollaboratorIDs:  []uuid.UUID{uuid.New()},
		ClientContactIDs: []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, domain.ErrInvalidClientContact) {
		t.Fatalf("err = %v, want ErrInvalidClientContact", err)
	}
}

type contactRepo struct {
	noopRepo
	contacts []ports.ClientContactSnapshot
	saved    domain.Mission
}

func (r *contactRepo) ListClientContacts(context.Context, kernel.TenantID, uuid.UUID) ([]ports.ClientContactSnapshot, error) {
	return r.contacts, nil
}

func (r *contactRepo) SaveMission(_ context.Context, m domain.Mission) error {
	r.saved = m
	return nil
}

func (r *contactRepo) CreateMissionWithRelations(_ context.Context, m domain.Mission, _, _ []uuid.UUID) error {
	r.saved = m
	return nil
}

func TestCreate_bindsClientContactIDs(t *testing.T) {
	contactID := uuid.New()
	repo := &contactRepo{
		contacts: []ports.ClientContactSnapshot{{
			ID: contactID, Prenom: "Marie", Nom: "Dupont", Email: "marie@acme.test",
		}},
	}
	svc := NewService(repo, nil, nil, nil)
	m, err := svc.Create(context.Background(), ports.CreateMissionCommand{
		TenantID:         kernel.NewTenantID(uuid.New()),
		ClientID:         uuid.New(),
		StartDate:        time.Now().UTC(),
		RateUnit:         "tjm",
		CollaboratorIDs:  []uuid.UUID{uuid.New()},
		ClientContactIDs: []uuid.UUID{contactID},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(m.ClientContactIDs) != 1 || m.ClientContactIDs[0] != contactID {
		t.Fatalf("contact ids = %v", m.ClientContactIDs)
	}
	if m.ClientContact != "Marie Dupont" {
		t.Fatalf("label = %q", m.ClientContact)
	}
}

type missionStoreRepo struct {
	noopRepo
	mission  domain.Mission
	contacts []ports.ClientContactSnapshot
}

func (r *missionStoreRepo) GetMission(context.Context, kernel.TenantID, uuid.UUID) (domain.Mission, error) {
	return r.mission, nil
}

func (r *missionStoreRepo) SaveMission(_ context.Context, m domain.Mission) error {
	r.mission = m
	return nil
}

func (r *missionStoreRepo) ListMissionCollaborators(context.Context, kernel.TenantID, uuid.UUID) ([]ports.MissionCollaborator, error) {
	return []ports.MissionCollaborator{}, nil
}

func (r *missionStoreRepo) ListMissionApplications(context.Context, kernel.TenantID, uuid.UUID) ([]ports.MissionApplication, error) {
	return []ports.MissionApplication{}, nil
}

func (r *missionStoreRepo) ListClientContacts(context.Context, kernel.TenantID, uuid.UUID) ([]ports.ClientContactSnapshot, error) {
	return r.contacts, nil
}

func TestUpdate_omittedContactIDsPreserves(t *testing.T) {
	contactID := uuid.New()
	repo := &missionStoreRepo{
		mission: domain.Mission{
			ID:               uuid.New(),
			TenantID:         kernel.NewTenantID(uuid.New()),
			ClientID:         uuid.New(),
			Title:            "Old",
			RateUnit:         domain.RateUnitTJM,
			ClientContact:    "Marie Dupont",
			ClientContactIDs: []uuid.UUID{contactID},
		},
		contacts: []ports.ClientContactSnapshot{{ID: contactID, Prenom: "Marie", Nom: "Dupont"}},
	}
	svc := NewService(repo, nil, nil, nil)
	_, err := svc.Update(context.Background(), ports.UpdateMissionCommand{
		TenantID:  repo.mission.TenantID,
		MissionID: repo.mission.ID,
		Title:     "New",
		RateUnit:  "tjm",
		TJMAmount: 1000,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(repo.mission.ClientContactIDs) != 1 || repo.mission.ClientContactIDs[0] != contactID {
		t.Fatalf("contacts cleared unexpectedly: %v", repo.mission.ClientContactIDs)
	}
}

func TestUpdate_emptyContactIDsClears(t *testing.T) {
	contactID := uuid.New()
	empty := []uuid.UUID{}
	repo := &missionStoreRepo{
		mission: domain.Mission{
			ID:               uuid.New(),
			TenantID:         kernel.NewTenantID(uuid.New()),
			ClientID:         uuid.New(),
			RateUnit:         domain.RateUnitTJM,
			ClientContact:    "Marie Dupont",
			ClientContactIDs: []uuid.UUID{contactID},
		},
		contacts: []ports.ClientContactSnapshot{{ID: contactID, Prenom: "Marie", Nom: "Dupont"}},
	}
	svc := NewService(repo, nil, nil, nil)
	detail, err := svc.Update(context.Background(), ports.UpdateMissionCommand{
		TenantID:         repo.mission.TenantID,
		MissionID:        repo.mission.ID,
		RateUnit:         "tjm",
		ClientContactIDs: &empty,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(repo.mission.ClientContactIDs) != 0 {
		t.Fatalf("expected cleared ids, got %v", repo.mission.ClientContactIDs)
	}
	if detail.ClientContact != "" || len(detail.ClientContacts) != 0 {
		t.Fatalf("expected empty detail contacts, got %+v", detail)
	}
}

func TestGetDetail_noLegacyFallbackWhenIDsPresent(t *testing.T) {
	repo := &missionStoreRepo{
		mission: domain.Mission{
			ID:               uuid.New(),
			TenantID:         kernel.NewTenantID(uuid.New()),
			ClientID:         uuid.New(),
			RateUnit:         domain.RateUnitTJM,
			ClientContact:    "Ghost Contact",
			ClientContactIDs: []uuid.UUID{uuid.New()},
		},
		contacts: nil,
	}
	svc := NewService(repo, nil, nil, nil)
	detail, err := svc.GetDetail(context.Background(), repo.mission.TenantID, repo.mission.ID)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.ClientContact != "" {
		t.Fatalf("ghost label = %q", detail.ClientContact)
	}
}

type capturingCalendar struct {
	lastCountry string
}

func (c *capturingCalendar) IsHolidayOrLeave(_ context.Context, _ kernel.TenantID, _ uuid.UUID, _ time.Time, countryCode string) (bool, error) {
	c.lastCountry = countryCode
	return false, nil
}

type paysRepo struct {
	noopRepo
	pays string
	err  error
}

func (r *paysRepo) CreateMissionWithRelations(context.Context, domain.Mission, []uuid.UUID, []uuid.UUID) error {
	return nil
}

func (r *paysRepo) GetClientPays(context.Context, kernel.TenantID, uuid.UUID) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return r.pays, nil
}

func TestCreate_usesClientPaysForPrefill(t *testing.T) {
	cal := &capturingCalendar{}
	repo := &paysRepo{pays: "be"}
	svc := NewService(repo, &fakeFeeder{}, nil, cal)
	end := time.Now().UTC().AddDate(0, 0, 14)
	_, err := svc.Create(context.Background(), ports.CreateMissionCommand{
		TenantID:        kernel.NewTenantID(uuid.New()),
		ClientID:        uuid.New(),
		StartDate:       time.Now().UTC(),
		EndDate:         &end,
		RateUnit:        "tjm",
		CollaboratorIDs: []uuid.UUID{uuid.New()},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if cal.lastCountry != "BE" {
		t.Fatalf("country = %q, want BE from client.pays", cal.lastCountry)
	}
}

func TestResolveClientCountry_defaultsAndAliases(t *testing.T) {
	cases := []struct {
		name string
		pays string
		err  error
		want string
	}{
		{name: "md alias", pays: "md", want: "MG"},
		{name: "unsupported", pays: "DE", want: "FR"},
		{name: "empty", pays: "", want: "FR"},
		{name: "spaces and casing", pays: "  Be ", want: "BE"},
		{name: "lookup error", err: errors.New("db down"), want: "FR"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &service{repo: &paysRepo{pays: tc.pays, err: tc.err}}
			got := svc.resolveClientCountry(context.Background(), kernel.NewTenantID(uuid.New()), uuid.New())
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

type appsRepo struct {
	noopRepo
	validIDs map[uuid.UUID]struct{}
	saved    []uuid.UUID
	mission  domain.Mission
}

func (r *appsRepo) ValidateApplicationIDs(_ context.Context, _ kernel.TenantID, ids []uuid.UUID, _ uuid.UUID) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]struct{})
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		if _, ok := r.validIDs[id]; !ok {
			continue
		}
		out = append(out, id)
	}
	return out, nil
}

func (r *appsRepo) CreateMissionWithRelations(_ context.Context, m domain.Mission, _, applicationIDs []uuid.UUID) error {
	r.mission = m
	r.saved = append([]uuid.UUID{}, applicationIDs...)
	return nil
}

func (r *appsRepo) SaveMissionApplications(_ context.Context, _ kernel.TenantID, _ uuid.UUID, ids []uuid.UUID) error {
	r.saved = append([]uuid.UUID{}, ids...)
	return nil
}

func (r *appsRepo) GetMission(context.Context, kernel.TenantID, uuid.UUID) (domain.Mission, error) {
	return r.mission, nil
}

func (r *appsRepo) ListMissionCollaborators(context.Context, kernel.TenantID, uuid.UUID) ([]ports.MissionCollaborator, error) {
	return []ports.MissionCollaborator{}, nil
}

func (r *appsRepo) ListMissionApplications(context.Context, kernel.TenantID, uuid.UUID) ([]ports.MissionApplication, error) {
	out := make([]ports.MissionApplication, 0, len(r.saved))
	for _, id := range r.saved {
		out = append(out, ports.MissionApplication{ApplicationID: id, Libelle: id.String(), Active: true})
	}
	return out, nil
}

func TestCreate_rejectsUnknownApplication(t *testing.T) {
	svc := NewService(&appsRepo{validIDs: map[uuid.UUID]struct{}{}}, nil, nil, nil)
	_, err := svc.Create(context.Background(), ports.CreateMissionCommand{
		TenantID:        kernel.NewTenantID(uuid.New()),
		ClientID:        uuid.New(),
		StartDate:       time.Now().UTC(),
		RateUnit:        "tjm",
		CollaboratorIDs: []uuid.UUID{uuid.New()},
		ApplicationIDs:  []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, domain.ErrInvalidApplication) {
		t.Fatalf("err = %v, want ErrInvalidApplication", err)
	}
}

func TestCreate_bindsOptionalApplications(t *testing.T) {
	appID := uuid.New()
	repo := &appsRepo{validIDs: map[uuid.UUID]struct{}{appID: {}}}
	svc := NewService(repo, nil, nil, nil)
	_, err := svc.Create(context.Background(), ports.CreateMissionCommand{
		TenantID:        kernel.NewTenantID(uuid.New()),
		ClientID:        uuid.New(),
		StartDate:       time.Now().UTC(),
		RateUnit:        "tjm",
		CollaboratorIDs: []uuid.UUID{uuid.New()},
		ApplicationIDs:  []uuid.UUID{appID, appID},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0] != appID {
		t.Fatalf("saved apps = %v", repo.saved)
	}
}

func TestUpdateApplications_replacesAndAllowsEmpty(t *testing.T) {
	appID := uuid.New()
	repo := &appsRepo{
		validIDs: map[uuid.UUID]struct{}{appID: {}},
		mission: domain.Mission{
			ID:       uuid.New(),
			TenantID: kernel.NewTenantID(uuid.New()),
			ClientID: uuid.New(),
			RateUnit: domain.RateUnitTJM,
		},
		saved: []uuid.UUID{appID},
	}
	svc := NewService(repo, nil, nil, nil)
	detail, err := svc.UpdateApplications(context.Background(), ports.UpdateApplicationsCommand{
		TenantID:       repo.mission.TenantID,
		MissionID:      repo.mission.ID,
		ApplicationIDs: nil,
	})
	if err != nil {
		t.Fatalf("UpdateApplications: %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("expected clear, got %v", repo.saved)
	}
	if len(detail.Applications) != 0 {
		t.Fatalf("detail apps = %+v", detail.Applications)
	}
}

func TestUpdateApplications_rejectsUnknown(t *testing.T) {
	repo := &appsRepo{
		validIDs: map[uuid.UUID]struct{}{},
		mission: domain.Mission{
			ID:       uuid.New(),
			TenantID: kernel.NewTenantID(uuid.New()),
			ClientID: uuid.New(),
			RateUnit: domain.RateUnitTJM,
		},
	}
	svc := NewService(repo, nil, nil, nil)
	_, err := svc.UpdateApplications(context.Background(), ports.UpdateApplicationsCommand{
		TenantID:       repo.mission.TenantID,
		MissionID:      repo.mission.ID,
		ApplicationIDs: []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, domain.ErrInvalidApplication) {
		t.Fatalf("err = %v, want ErrInvalidApplication", err)
	}
}

func TestGetDetail_includesApplications(t *testing.T) {
	appID := uuid.New()
	repo := &appsRepo{
		validIDs: map[uuid.UUID]struct{}{appID: {}},
		mission: domain.Mission{
			ID:       uuid.New(),
			TenantID: kernel.NewTenantID(uuid.New()),
			ClientID: uuid.New(),
			RateUnit: domain.RateUnitTJM,
		},
		saved: []uuid.UUID{appID},
	}
	svc := NewService(repo, nil, nil, nil)
	detail, err := svc.GetDetail(context.Background(), repo.mission.TenantID, repo.mission.ID)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if len(detail.Applications) != 1 || detail.Applications[0].ApplicationID != appID {
		t.Fatalf("applications = %+v", detail.Applications)
	}
}

type noopRepo struct{}

func (noopRepo) SaveMission(context.Context, domain.Mission) error { return nil }
func (noopRepo) CreateMissionWithRelations(context.Context, domain.Mission, []uuid.UUID, []uuid.UUID) error {
	return nil
}
func (noopRepo) GetMission(context.Context, kernel.TenantID, uuid.UUID) (domain.Mission, error) {
	return domain.Mission{}, domain.ErrMissionNotFound
}
func (noopRepo) ListMissions(context.Context, kernel.TenantID) ([]domain.Mission, error) {
	return nil, nil
}
func (noopRepo) ListMissionSummaries(context.Context, kernel.TenantID) ([]ports.MissionSummary, error) {
	return nil, nil
}
func (noopRepo) ListMissionCollaborators(context.Context, kernel.TenantID, uuid.UUID) ([]ports.MissionCollaborator, error) {
	return nil, nil
}
func (noopRepo) SaveMissionCollaborators(context.Context, kernel.TenantID, uuid.UUID, []uuid.UUID) error {
	return nil
}
func (noopRepo) ListMissionApplications(context.Context, kernel.TenantID, uuid.UUID) ([]ports.MissionApplication, error) {
	return nil, nil
}
func (noopRepo) SaveMissionApplications(context.Context, kernel.TenantID, uuid.UUID, []uuid.UUID) error {
	return nil
}
func (noopRepo) ValidateApplicationIDs(_ context.Context, _ kernel.TenantID, ids []uuid.UUID, _ uuid.UUID) ([]uuid.UUID, error) {
	return ids, nil
}
func (noopRepo) GetClientName(context.Context, kernel.TenantID, uuid.UUID) (string, error) {
	return "", nil
}
func (noopRepo) GetClientPays(context.Context, kernel.TenantID, uuid.UUID) (string, error) {
	return "FR", nil
}
func (noopRepo) ListClientContacts(context.Context, kernel.TenantID, uuid.UUID) ([]ports.ClientContactSnapshot, error) {
	return nil, nil
}
func (noopRepo) PurgeClientContactsFromMissions(context.Context, kernel.TenantID, uuid.UUID, []uuid.UUID) error {
	return nil
}
