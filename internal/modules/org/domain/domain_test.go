package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFormatSocieteAddress(t *testing.T) {
	got := FormatSocieteAddress(Societe{
		Adresse:       "Rue de la Résistance",
		AdresseNumero: "92",
		AdresseBoite:  "A",
		CodePostal:    "4100",
		Ville:         "Seraing",
		Pays:          "BE",
	})
	require.Equal(t, "Rue de la Résistance 92 / A, 4100 Seraing, Belgique", got)

	gotFR := FormatSocieteAddress(Societe{
		Adresse:    "1 rue de la Paix",
		CodePostal: "75002",
		Ville:      "Paris",
		Pays:       "FR",
	})
	require.Equal(t, "1 rue de la Paix, 75002 Paris, France", gotFR)

	gotMA := FormatSocieteAddress(Societe{
		Adresse:    "Bd Zerktouni",
		CodePostal: "20000",
		Ville:      "Casablanca",
		Pays:       "ma",
	})
	require.Equal(t, "Bd Zerktouni, 20000 Casablanca, Maroc", gotMA)

	gotCA := FormatSocieteAddress(Societe{
		Adresse:    "Rue Sainte-Catherine",
		CodePostal: "H3B 1A7",
		Ville:      "Montréal",
		Pays:       "CA",
	})
	require.Equal(t, "Rue Sainte-Catherine, H3B 1A7 Montréal, Canada", gotCA)
}

func TestFormatClientAddress(t *testing.T) {
	got := FormatClientAddress(Client{
		Adresse:       "Avenue Louise",
		AdresseNumero: "100",
		CodePostal:    "1050",
		Ville:         "Ixelles",
		Pays:          "BE",
	})
	require.Equal(t, "Avenue Louise 100, 1050 Ixelles, Belgique", got)

	empty := FormatClientAddress(Client{RaisonSociale: "Acme"})
	require.Empty(t, empty)
}

func TestNormalizeClientContacts(t *testing.T) {
	existing := uuid.New()
	got := NormalizeClientContacts([]ClientContact{
		{ID: existing, Prenom: "  Marie ", Nom: "Dupont", Role: " DSI "},
		{Prenom: "Paul", Email: "paul@test.com"},
	})
	require.Len(t, got, 2)
	require.Equal(t, existing, got[0].ID)
	require.Equal(t, "Marie", got[0].Prenom)
	require.Equal(t, "DSI", got[0].Role)
	require.NotEqual(t, uuid.Nil, got[1].ID)
	require.Equal(t, "Paul", ClientContactDisplayName(got[1]))
}

func TestEnsureClientContactIDs_reportsChange(t *testing.T) {
	existing := uuid.New()
	_, changed := EnsureClientContactIDs([]ClientContact{{ID: existing, Prenom: "A"}})
	require.False(t, changed)
	_, changed = EnsureClientContactIDs([]ClientContact{{Prenom: "B"}})
	require.True(t, changed)
}

func TestSanitizeClientContacts_keepsNilID(t *testing.T) {
	got := SanitizeClientContacts([]ClientContact{{Prenom: "  A ", Nom: "B"}})
	require.Len(t, got, 1)
	require.Equal(t, uuid.Nil, got[0].ID)
	require.Equal(t, "A", got[0].Prenom)
}

func TestLoginValid(t *testing.T) {
	tests := []struct {
		name  string
		login string
		ok    bool
	}{
		{"valid with prefix", "ABC_jean", true},
		{"valid without prefix", "olivier", true},
		{"valid dotted", "jean.dupont", true},
		{"valid hyphen", "jean-dupont", true},
		{"valid mixed case", "JeanDupont", true},
		{"valid seed style", "ADM_admin", true},
		{"invalid too short", "ab", false},
		{"invalid starts with digit", "1olivier", false},
		{"invalid space", "jean dupont", false},
		{"invalid special char", "jean@dupont", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLogin(tt.login)
			if tt.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrInvalidLogin)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name string
		pwd  string
		ok   bool
	}{
		{"strong", "Admin123!", true},
		{"too short", "Ab1!", false},
		{"no upper", "admin123!", false},
		{"no lower", "ADMIN123!", false},
		{"no digit", "AdminPass!", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.pwd)
			if tt.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, ErrWeakPassword)
			}
		})
	}
}

func TestActivationPeriodExpired(t *testing.T) {
	exp := time.Now().Add(-24 * time.Hour)
	period := ActivationPeriod{
		Activation: time.Now().Add(-48 * time.Hour),
		Expiration: &exp,
	}
	require.False(t, period.IsActive(time.Now()))
}

func TestApplication_HasSharesAndDedupe(t *testing.T) {
	require.False(t, Application{}.HasShares())
	require.True(t, Application{ServiceIDs: []uuid.UUID{uuid.New()}}.HasShares())
	require.True(t, Application{SiteIDs: []uuid.UUID{uuid.New()}}.HasShares())
	require.True(t, Application{EquipeIDs: []uuid.UUID{uuid.New()}}.HasShares())

	a, b := uuid.New(), uuid.New()
	got := DedupeUUIDs([]uuid.UUID{a, uuid.Nil, a, b})
	require.Equal(t, []uuid.UUID{a, b}, got)
}
