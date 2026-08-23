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

type brandingRepo struct {
	refreshUserRepo
	societe      domain.Societe
	savedLogo    []byte
	savedType    string
	updated      *domain.Societe
	logoByTenant map[uuid.UUID]struct {
		content     []byte
		contentType string
	}
}

func (r *brandingRepo) GetSociete(context.Context, kernel.TenantID, uuid.UUID) (domain.Societe, error) {
	return r.societe, nil
}

func (r *brandingRepo) UpdateSociete(_ context.Context, s domain.Societe) error {
	r.updated = &s
	r.societe = s
	return nil
}

func (r *brandingRepo) SaveSocieteLogo(_ context.Context, tenant kernel.TenantID, _ uuid.UUID, content []byte, contentType string) error {
	r.savedLogo = append([]byte(nil), content...)
	r.savedType = contentType
	if r.logoByTenant == nil {
		r.logoByTenant = map[uuid.UUID]struct {
			content     []byte
			contentType string
		}{}
	}
	r.logoByTenant[tenant.UUID()] = struct {
		content     []byte
		contentType string
	}{content: append([]byte(nil), content...), contentType: contentType}
	return nil
}

func (r *brandingRepo) GetTenantLogo(_ context.Context, tenant kernel.TenantID) ([]byte, string, error) {
	if r.logoByTenant == nil {
		return nil, "", domain.ErrLogoNotFound
	}
	entry, ok := r.logoByTenant[tenant.UUID()]
	if !ok {
		return nil, "", domain.ErrLogoNotFound
	}
	return entry.content, entry.contentType, nil
}

func TestUpdateSocieteBranding_PersistsLogoContent(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	societeID := uuid.New()
	repo := &brandingRepo{
		societe: domain.Societe{
			ID:            societeID,
			TenantID:      tenant,
			RaisonSociale: "Avant",
		},
	}
	svc := NewOrganizationService(repo, nil)

	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x01}
	got, err := svc.UpdateSocieteBranding(context.Background(), ports.UpdateSocieteBrandingCommand{
		TenantID:        tenant,
		SocieteID:       societeID,
		RaisonSociale:   "LL-IT",
		Logo:            "/api/v1/branding/logo/" + tenant.UUID().String(),
		LogoContent:     png,
		LogoContentType: "image/png",
		Adresse:         "Rue de la Résistance",
		AdresseNumero:   "92",
		AdresseBoite:    "A",
		CodePostal:      "4100",
		Ville:           "Seraing",
		Pays:            "BE",
		Siret:           "BE1007132489",
	})
	require.NoError(t, err)
	require.Equal(t, "LL-IT", got.RaisonSociale)
	require.Equal(t, "/api/v1/branding/logo/"+tenant.UUID().String(), got.Logo)
	require.Equal(t, "BE", got.Pays)
	require.Equal(t, "92", got.AdresseNumero)
	require.Equal(t, "4100", got.CodePostal)
	require.Equal(t, "Seraing", got.Ville)
	require.NotNil(t, repo.updated)
	require.Equal(t, png, repo.savedLogo)
	require.Equal(t, "image/png", repo.savedType)

	content, ct, err := svc.GetTenantLogo(context.Background(), tenant)
	require.NoError(t, err)
	require.Equal(t, "image/png", ct)
	require.Equal(t, png, content)
}

func TestGetTenantLogo_NotFound(t *testing.T) {
	svc := NewOrganizationService(&brandingRepo{}, nil)
	_, _, err := svc.GetTenantLogo(context.Background(), kernel.NewTenantID(uuid.New()))
	require.ErrorIs(t, err, domain.ErrLogoNotFound)
}

func TestUpdateSocieteBranding_AcceptsNewCountries(t *testing.T) {
	tenant := kernel.NewTenantID(uuid.New())
	societeID := uuid.New()
	repo := &brandingRepo{
		societe: domain.Societe{
			ID:            societeID,
			TenantID:      tenant,
			RaisonSociale: "Avant",
			Pays:          "FR",
		},
	}
	svc := NewOrganizationService(repo, nil)

	for _, country := range []string{"MA", "TN", "MG", "CA"} {
		got, err := svc.UpdateSocieteBranding(context.Background(), ports.UpdateSocieteBrandingCommand{
			TenantID:      tenant,
			SocieteID:     societeID,
			RaisonSociale: "Société " + country,
			Pays:          country,
			Siret:         "REG-" + country,
		})
		require.NoError(t, err)
		require.Equal(t, country, got.Pays)
	}
}
