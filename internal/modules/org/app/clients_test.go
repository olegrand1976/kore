package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kore/kore/internal/modules/org/domain"
	"github.com/kore/kore/internal/modules/org/ports"
	"github.com/kore/kore/pkg/kernel"
	"github.com/stretchr/testify/require"
)

type clientRepoFake struct {
	refreshUserRepo
	clients map[uuid.UUID]domain.Client
}

func (r *clientRepoFake) SaveClient(_ context.Context, c domain.Client) error {
	if r.clients == nil {
		r.clients = map[uuid.UUID]domain.Client{}
	}
	r.clients[c.ID] = c
	return nil
}

func (r *clientRepoFake) UpdateClient(_ context.Context, c domain.Client) error {
	if r.clients == nil {
		return domain.ErrClientNotFound
	}
	if _, ok := r.clients[c.ID]; !ok {
		return domain.ErrClientNotFound
	}
	r.clients[c.ID] = c
	return nil
}

func (r *clientRepoFake) UpdateClientContacts(_ context.Context, _ kernel.TenantID, id uuid.UUID, contacts []domain.ClientContact) error {
	if r.clients == nil {
		return domain.ErrClientNotFound
	}
	c, ok := r.clients[id]
	if !ok {
		return domain.ErrClientNotFound
	}
	c.Contacts = contacts
	r.clients[id] = c
	return nil
}

func (r *clientRepoFake) GetClient(_ context.Context, _ kernel.TenantID, id uuid.UUID) (domain.Client, error) {
	if r.clients == nil {
		return domain.Client{}, domain.ErrClientNotFound
	}
	c, ok := r.clients[id]
	if !ok {
		return domain.Client{}, domain.ErrClientNotFound
	}
	return c, nil
}

func (r *clientRepoFake) ListClients(context.Context, kernel.TenantID) ([]domain.Client, error) {
	out := make([]domain.Client, 0, len(r.clients))
	for _, c := range r.clients {
		out = append(out, c)
	}
	return out, nil
}

func TestCreateClient_withBillingFields(t *testing.T) {
	repo := &clientRepoFake{}
	svc := NewClientService(repo)
	tenant := kernel.NewTenantID(uuid.New())

	got, err := svc.CreateClient(context.Background(), ports.CreateClientCommand{
		TenantID:      tenant,
		RaisonSociale: "  Acme SA  ",
		TVA:           " BE0123456789 ",
		Pays:          "be",
		Adresse:       "Rue du Midi",
		AdresseNumero: "10",
		AdresseBoite:  "B",
		CodePostal:    "1000",
		Ville:         "Bruxelles",
		Siret:         "0123456789",
	})
	require.NoError(t, err)
	require.Equal(t, "Acme SA", got.RaisonSociale)
	require.Equal(t, "BE0123456789", got.TVA)
	require.Equal(t, "BE", got.Pays)
	require.Equal(t, "Rue du Midi", got.Adresse)
	require.Equal(t, "10", got.AdresseNumero)
	require.Equal(t, "B", got.AdresseBoite)
	require.Equal(t, "1000", got.CodePostal)
	require.Equal(t, "Bruxelles", got.Ville)
	require.Equal(t, "0123456789", got.Siret)
	require.Contains(t, repo.clients, got.ID)
}

func TestCreateClient_rejectsEmptyName(t *testing.T) {
	svc := NewClientService(&clientRepoFake{})
	_, err := svc.CreateClient(context.Background(), ports.CreateClientCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		RaisonSociale: "   ",
	})
	require.ErrorIs(t, err, domain.ErrInvalidClientName)
}

func TestCreateClient_rejectsInvalidPays(t *testing.T) {
	svc := NewClientService(&clientRepoFake{})
	_, err := svc.CreateClient(context.Background(), ports.CreateClientCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		RaisonSociale: "Acme",
		Pays:          "DE",
	})
	require.ErrorIs(t, err, domain.ErrInvalidPays)
}

func TestCreateClient_emptyPaysDefaultsToFR(t *testing.T) {
	svc := NewClientService(&clientRepoFake{})
	got, err := svc.CreateClient(context.Background(), ports.CreateClientCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		RaisonSociale: "Acme",
	})
	require.NoError(t, err)
	require.Equal(t, "FR", got.Pays)
}

func TestUpdateClient_billingRoundTrip(t *testing.T) {
	repo := &clientRepoFake{}
	svc := NewClientService(repo)
	tenant := kernel.NewTenantID(uuid.New())

	created, err := svc.CreateClient(context.Background(), ports.CreateClientCommand{
		TenantID:      tenant,
		RaisonSociale: "Acme",
		Pays:          "FR",
	})
	require.NoError(t, err)

	updated, err := svc.UpdateClient(context.Background(), ports.UpdateClientCommand{
		TenantID:      tenant,
		ClientID:      created.ID,
		RaisonSociale: "Acme Updated",
		TVA:           "FR123",
		Pays:          "MA",
		Adresse:       "Bd Zerktouni",
		CodePostal:    "20000",
		Ville:         "Casablanca",
		Siret:         "123456789012345",
	})
	require.NoError(t, err)
	require.Equal(t, "Acme Updated", updated.RaisonSociale)
	require.Equal(t, "MA", updated.Pays)
	require.Equal(t, "Casablanca", updated.Ville)
	require.Equal(t, "123456789012345", updated.Siret)
}

func TestUpdateClient_notFound(t *testing.T) {
	svc := NewClientService(&clientRepoFake{})
	_, err := svc.UpdateClient(context.Background(), ports.UpdateClientCommand{
		TenantID:      kernel.NewTenantID(uuid.New()),
		ClientID:      uuid.New(),
		RaisonSociale: "X",
	})
	require.ErrorIs(t, err, domain.ErrClientNotFound)
}

func TestReplaceClientContacts_assignsIDs(t *testing.T) {
	repo := &clientRepoFake{}
	svc := NewClientService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	created, err := svc.CreateClient(context.Background(), ports.CreateClientCommand{
		TenantID:      tenant,
		RaisonSociale: "Acme",
	})
	require.NoError(t, err)

	got, err := svc.ReplaceClientContacts(context.Background(), ports.ReplaceClientContactsCommand{
		TenantID: tenant,
		ClientID: created.ID,
		Contacts: []domain.ClientContact{{
			Prenom: "Marie",
			Nom:    "Dupont",
			Email:  "marie@acme.test",
			Role:   "DSI",
		}},
	})
	require.NoError(t, err)
	require.Len(t, got.Contacts, 1)
	require.NotEqual(t, uuid.Nil, got.Contacts[0].ID)
	require.Equal(t, "DSI", got.Contacts[0].Role)
}

type missionCleanerFake struct {
	calls []struct {
		clientID uuid.UUID
		removed  []uuid.UUID
	}
}

func (f *missionCleanerFake) PurgeRemovedClientContacts(_ context.Context, _ kernel.TenantID, clientID uuid.UUID, removedIDs []uuid.UUID) error {
	f.calls = append(f.calls, struct {
		clientID uuid.UUID
		removed  []uuid.UUID
	}{clientID: clientID, removed: append([]uuid.UUID{}, removedIDs...)})
	return nil
}

func TestReplaceClientContacts_purgesRemovedFromMissions(t *testing.T) {
	repo := &clientRepoFake{}
	cleaner := &missionCleanerFake{}
	svc := NewClientService(repo, WithMissionContactCleaner(cleaner))
	tenant := kernel.NewTenantID(uuid.New())
	created, err := svc.CreateClient(context.Background(), ports.CreateClientCommand{
		TenantID:      tenant,
		RaisonSociale: "Acme",
	})
	require.NoError(t, err)

	keepID := uuid.New()
	dropID := uuid.New()
	_, err = svc.ReplaceClientContacts(context.Background(), ports.ReplaceClientContactsCommand{
		TenantID: tenant,
		ClientID: created.ID,
		Contacts: []domain.ClientContact{
			{ID: keepID, Prenom: "Keep"},
			{ID: dropID, Prenom: "Drop"},
		},
	})
	require.NoError(t, err)
	require.Empty(t, cleaner.calls)

	_, err = svc.ReplaceClientContacts(context.Background(), ports.ReplaceClientContactsCommand{
		TenantID: tenant,
		ClientID: created.ID,
		Contacts: []domain.ClientContact{
			{ID: keepID, Prenom: "Keep"},
		},
	})
	require.NoError(t, err)
	require.Len(t, cleaner.calls, 1)
	require.Equal(t, created.ID, cleaner.calls[0].clientID)
	require.Equal(t, []uuid.UUID{dropID}, cleaner.calls[0].removed)
}

func TestGetClient_doesNotPersistMissingIDs(t *testing.T) {
	repo := &clientRepoFake{}
	svc := NewClientService(repo)
	tenant := kernel.NewTenantID(uuid.New())
	id := uuid.New()
	repo.clients = map[uuid.UUID]domain.Client{
		id: {ID: id, TenantID: tenant, RaisonSociale: "Acme", Contacts: []domain.ClientContact{{Prenom: "Legacy"}}},
	}
	got, err := svc.GetClient(context.Background(), tenant, id)
	require.NoError(t, err)
	require.Equal(t, uuid.Nil, got.Contacts[0].ID)
	require.Equal(t, uuid.Nil, repo.clients[id].Contacts[0].ID)
}
