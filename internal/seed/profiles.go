package seed

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/kore/kore/pkg/kernel"
)

type userMetaSpec struct {
	login      string
	typeCompte string
	langue     string
	craRequis  bool
	salarieETT bool
}

func demoUserMetaSpecs() []userMetaSpec {
	return []userMetaSpec{
		{AdminLogin, "Interne", "fr", false, false},
		{ManagerLogin, "Interne", "fr", true, false},
		{CommercialLogin, "Interne", "fr", false, false},
		{CollabLogin, "Interne", "fr", true, false},
		{Collab2Login, "Interne", "fr", true, false},
		{PrestaLogin, "Prestataire", "fr", true, true},
		{ClientUserLogin, "Client", "fr", false, false},
	}
}

func (r *Runner) enrichUserProfiles(ctx context.Context, tenant kernel.TenantID) error {
	for _, spec := range demoUserMetaSpecs() {
		if err := r.patchUserMeta(ctx, tenant, spec.login, spec.typeCompte, spec.langue, spec.craRequis, spec.salarieETT); err != nil {
			return err
		}
	}
	log.Println("seed: métadonnées utilisateurs (type, langue, CRA) enrichies pour tous les profils")
	return nil
}

func (r *Runner) enrichOrgStructure(ctx context.Context, tenant kernel.TenantID, oc orgContext) error {
	if _, err := r.deps.Pool.Exec(ctx, `
		UPDATE org.services
		SET commercial_id = $3, assistante_id = $4
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), DemoServiceID, oc.commercialID, oc.adminID); err != nil {
		return err
	}

	if _, err := r.deps.Pool.Exec(ctx, `
		UPDATE org.applications
		SET proprietaire = $3, chef_utilisateur_id = $4, uo_activee = TRUE
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), DemoAppID, DemoClientName, oc.collabID); err != nil {
		return err
	}

	if _, err := r.deps.Pool.Exec(ctx, `
		UPDATE org.applications
		SET proprietaire = $3, chef_utilisateur_id = $4, uo_activee = TRUE, mode_facturation = 'forfait'
		WHERE tenant_id = $1 AND id = $2
	`, tenant.UUID(), DemoApp2ID, DemoClient2Name, oc.prestaID); err != nil {
		return err
	}

	if err := r.ensureClientContacts(ctx, tenant, DemoClientName, clientContactsACME); err != nil {
		return err
	}
	if err := r.ensureClientContacts(ctx, tenant, DemoClient2Name, clientContactsGlobex); err != nil {
		return err
	}

	log.Println("seed: structure org enrichie (commercial, chefs de projet, contacts clients)")
	return nil
}

const (
	clientContactsACME = `[{"nom":"Dupont","prenom":"Marie","email":"marie.dupont@acme.test","role":"DSI","telephone":"+33 1 42 00 00 01"},{"nom":"Martin","prenom":"Paul","email":"paul.martin@acme.test","role":"MOA Portail","telephone":"+33 1 42 00 00 02"}]`
	clientContactsGlobex = `[{"nom":"Schmidt","prenom":"Anna","email":"anna.schmidt@globex.test","role":"Directrice SI","telephone":"+33 4 78 00 00 03"}]`
)

func (r *Runner) ensureClientContacts(ctx context.Context, tenant kernel.TenantID, name, contactsJSON string) error {
	_, err := r.deps.Pool.Exec(ctx, `
		UPDATE org.clients
		SET contacts = $3::jsonb
		WHERE tenant_id = $1 AND raison_sociale = $2
		  AND (contacts IS NULL OR contacts = '[]'::jsonb)
	`, tenant.UUID(), name, contactsJSON)
	return err
}

func (r *Runner) patchUserMeta(
	ctx context.Context,
	tenant kernel.TenantID,
	login, typeCompte, langue string,
	craRequis, salarieETT bool,
) error {
	_, err := r.deps.Pool.Exec(ctx, `
		UPDATE org.users
		SET type_compte = $3, langue = $4, cra_requis = $5, salarie_ett = $6
		WHERE tenant_id = $1 AND login = $2
	`, tenant.UUID(), login, typeCompte, langue, craRequis, salarieETT)
	return err
}

func assigneeForApp(oc orgContext, appID uuid.UUID) uuid.UUID {
	switch appID {
	case oc.app2ID:
		return oc.prestaID
	case oc.app3ID:
		if id := oc.userID(CollabIntegLogin); id != uuid.Nil {
			return id
		}
		return oc.prestaID
	default:
		return oc.collabID
	}
}
