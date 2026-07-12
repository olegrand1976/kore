# Brique 00 — Organisation & Identity

> Première brique métier. Fournit le référentiel organisationnel, les comptes, l'authentification et le socle RBAC dont dépendent toutes les autres briques.

## 1. Référence fonctionnelle

- Spec §4 (modèle organisationnel), §3 (acteurs, profils, matrice RBAC §3.3), §11 (entités Société, Site, Service, Application, Équipe, Utilisateur, Client).
- Règles : RG-ORG-01 (login `XXX_nom`), RG-ORG-02 (Manager = Assistante), RG-SEC-01 (données privées), RG-SEC-02 (activation/expiration).
- Processus : PR-08.1 (mise en place initiale, 7 étapes).
- Fondations : [04-auth-rbac.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/04-auth-rbac.md), [01-architecture.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/01-architecture.md).

## 2. Périmètre de la brique et dépendances

**Inclus** : hiérarchie Société→Site→Service→Application→Équipe→Utilisateur ; comptes et authentification ; profils et permissions RBAC ; référentiel Client ; multi-tenant ; **configuration IdP SSO** par tenant (Phase 1, cf. [12-sso-federation.md](../foundation/12-sso-federation.md)).

**Hors brique** : workflow (01), CRA (02), logique métier des modules consommateurs.

**Dépend de** : foundation. Consomme `EntitlementReader` du [module 14](/home/olivier/ll-it-sc/projets/kore/technical/modules/14-abonnement-saas-stripe.md) pour le **plafond de sièges** : la création d'un utilisateur est refusée si `utilisateurs actifs ≥ seats` souscrits (`409 SEAT_LIMIT_REACHED`). **Consommée par** : toutes les autres briques (identité, tenant, RBAC, référentiel org).

```mermaid
flowchart LR
  Foundation --> Org["00 Organisation & Identity"]
  Org --> Autres["Toutes briques métier"]
```

## 3. Modèle de domaine

- **Agrégat Organisation** : `Societe` (racine tenant) -> `Site` -> `Service` -> `Application` -> `Equipe`.
- **Entité Utilisateur** : `login` (VO `Login` format `XXX_nom`), `profil` (VO `Profile`), `typeCompte` (Interne/Client/Prestataire), `craRequis`, `salarieETT`, période d'activation (VO `ActivationPeriod`).
- **Entité Client** : référentiel client (raison sociale, contacts, TVA).
- **Value objects** : `Login`, `Profile`, `ActivationPeriod`, `TenantID`.
- **Invariants** :
  - Un `Service` a un responsable (PR-08.1 étape 3) — ADMIN temporaire toléré.
  - `Application` avec TMA : budget par défaut obligatoire (RG-BUD-01, vérifié pleinement en brique 04/05).
  - `Login` respecte `XXX_nom` (RG-ORG-01).
  - Manager et Assistante = même profil socle (RG-ORG-02).

## 4. Ports

### Inbound (use cases)

```go
type OrganizationService interface {
    CreateSociete(ctx context.Context, cmd CreateSocieteCommand) (Societe, error)
    CreateSite(ctx context.Context, cmd CreateSiteCommand) (Site, error)
    CreateService(ctx context.Context, cmd CreateServiceCommand) (Service, error)
    CreateApplication(ctx context.Context, cmd CreateApplicationCommand) (Application, error)
    AssignServiceResponsible(ctx context.Context, cmd AssignResponsibleCommand) error
}

type UserService interface {
    CreateUser(ctx context.Context, cmd CreateUserCommand) (User, error)
    Authenticate(ctx context.Context, login, password string) (AuthResult, error)
    DeactivateUser(ctx context.Context, id UserID) error
}

type ClientService interface {
    CreateClient(ctx context.Context, cmd CreateClientCommand) (Client, error)
    ArchiveClient(ctx context.Context, id ClientID) error // refus si missions actives (RG-MISS)
}
```

### Outbound (repositories / gateways)

```go
type OrganizationRepository interface {
    SaveSociete(ctx context.Context, s Societe) error
    SaveSite(ctx context.Context, s Site) error
    // ... Service, Application, Equipe
    GetApplication(ctx context.Context, tenant TenantID, id ApplicationID) (Application, error)
}

type UserRepository interface {
    Save(ctx context.Context, u User) error
    FindByLogin(ctx context.Context, tenant TenantID, login string) (User, error)
    ExistsLogin(ctx context.Context, tenant TenantID, login string) (bool, error)
}

type PasswordHasher interface {
    Hash(plain string) (string, error)
    Verify(hash, plain string) bool
}

type TokenIssuer interface { // implémenté par platform/authx
    Issue(identity Identity) (access string, refresh string, err error)
}

type IdentityProviderRepository interface {
    Save(ctx context.Context, idp IdentityProvider) error
    FindByTenant(ctx context.Context, tenant TenantID) ([]IdentityProvider, error)
}

type UserIdentityRepository interface {
    Link(ctx context.Context, link UserIdentityLink) error
    FindBySubject(ctx context.Context, tenant TenantID, idpID IdPID, subject string) (UserIdentityLink, error)
}
```

## 5. Adapters

- **HTTP (chi)** : `internal/modules/org/adapters/http` — routes CRUD org, auth, clients.
- **PostgreSQL (sqlc)** : `internal/modules/org/adapters/postgres` — schéma `org`.
- **Gateways** : `PasswordHasher` (argon2id), `TokenIssuer` (platform/authx).

## 6. Contrat d'API

| Méthode | Chemin | Permission | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/login` | public | Authentification, pose cookies (via BFF) |
| POST | `/api/v1/auth/refresh` | public (refresh cookie) | Renouvellement token |
| POST | `/api/v1/auth/logout` | authentifié | Invalidation session |
| GET | `/api/v1/auth/oidc/authorize` | public | Redirection SSO (Phase 1) |
| POST | `/api/v1/auth/oidc/callback` | public | Callback OIDC → JWT Kore |
| GET | `/api/v1/admin/identity-providers` | Admin (L) | Liste IdP du tenant |
| PUT | `/api/v1/admin/identity-providers/{id}` | Admin (E) | Configurer IdP (Enterprise) |
| POST | `/api/v1/societes` | Admin (E) | Créer société |
| POST | `/api/v1/sites` | Admin (E) | Créer site |
| POST | `/api/v1/services` | Admin (E) | Créer service (responsable requis) |
| POST | `/api/v1/applications` | Admin (E) | Créer application |
| GET | `/api/v1/applications/{id}` | selon RBAC | Détail application |
| POST | `/api/v1/users` | Admin (E) | Créer compte (`XXX_nom`) |
| POST | `/api/v1/services/{id}/responsible` | Admin (E) | Affecter responsable |
| POST | `/api/v1/clients` | Resp. service / Commercial (E) | Créer client |

Erreurs clés : `409 LOGIN_ALREADY_EXISTS`, `422 INVALID_LOGIN_FORMAT`, `422 SERVICE_WITHOUT_RESPONSIBLE`, `401 INVALID_CREDENTIALS`, `403 ACCOUNT_EXPIRED`.

## 7. Schéma de données (schéma `org`)

| Table | Colonnes clés |
| --- | --- |
| `org.societes` | `id`, `tenant_id`, `raison_sociale`, `logo`, `devise`, `langue_defaut` |
| `org.sites` | `id`, `tenant_id`, `societe_id`, `libelle`, `pays`, `strategie_budget` |
| `org.services` | `id`, `tenant_id`, `site_id`, `type`, `responsable_id`, `commercial_id`, `assistante_id`, suppléants |
| `org.applications` | `id`, `tenant_id`, `service_id`, `proprietaire`, `mode_facturation`, `budget_defaut_id`, `uo_activee`, `chef_utilisateur_id` |
| `org.equipes` | `id`, `tenant_id`, `application_id`, `libelle`, `responsable_id` |
| `org.users` | `id`, `tenant_id`, `equipe_id`, `login`, `password_hash`, `profil`, `langue`, `cra_requis`, `type_compte`, `salarie_ett`, `date_activation`, `date_expiration` |
| `org.clients` | `id`, `tenant_id`, `raison_sociale`, `tva`, `contacts` |
| `authx.permissions` | `profile`, `module`, `action` (seed matrice §3.3) |
| `org.identity_providers` | `id`, `tenant_id`, `issuer`, `client_id`, `jwks_uri`, `scopes`, `enabled` (Phase 1) |
| `org.user_identities` | `id`, `tenant_id`, `user_id`, `idp_id`, `subject`, `email` (Phase 1) |

Contraintes : `UNIQUE (tenant_id, login)`, `CHECK` format login, index `(tenant_id, ...)`.

## 8. Mapping SOLID

| Principe | Application |
| --- | --- |
| SRP | `OrganizationService`, `UserService`, `ClientService` séparés par responsabilité ; hachage isolé dans `PasswordHasher`. |
| OCP | Nouveaux profils/permissions ajoutés par données (`authx.permissions`) sans modifier le code d'autorisation. |
| LSP | `UserRepository` postgres et mock interchangeables dans les tests d'`app`. |
| ISP | Ports fins (`PasswordHasher`, `TokenIssuer`) plutôt qu'un service auth monolithique. |
| DIP | `app` dépend de `UserRepository`/`PasswordHasher` (abstractions) ; implémentations câblées au composition root. |

## 9. Plan de tests unitaires

**Domaine** :
- `Login` valide/invalide (RG-ORG-01) — table-driven.
- `ActivationPeriod` : compte expiré rejeté (RG-SEC-02).
- Invariant service sans responsable.

**Application (mocks des ports)** :
- `CreateUser` : login dupliqué -> `LOGIN_ALREADY_EXISTS` ; format invalide -> erreur.
- `Authenticate` : mauvais mot de passe -> `INVALID_CREDENTIALS` ; compte expiré -> `ACCOUNT_EXPIRED` ; succès -> émission token (mock `TokenIssuer`).
- `ArchiveClient` : refus si dépendances (préparé pour brique SSII).
- Masquage données privées selon profil (RG-SEC-01).

**Intégration (testcontainers)** :
- Unicité `(tenant_id, login)` ; isolation multi-tenant (un tenant ne lit pas l'autre).

**Contrat HTTP** :
- 401/403 selon RBAC ; 422 format login ; login pose bien la réponse attendue.

Couverture cible : domaine > 90 %, app > 80 %.

## 10. Frontend Nuxt

| Élément | Détail |
| --- | --- |
| Pages | `login`, `admin/organisation` (arbre société/site/service/app), `admin/users`, `clients` |
| Composants | `OrgTree`, `UserForm`, `ClientForm` |
| Composables | `useAuth()`, `useOrganization()`, `useClients()` |
| Store Pinia | `auth`, `organization` |
| Routes BFF | `server/api/auth/*`, `server/api/organisation/*`, `server/api/users/*`, `server/api/clients/*` |
| Permissions UI | Menu admin visible profil Administrateur ; clients visibles Resp./Commercial |

## 11bis. SSO / Fédération d'identité (Phase 1)

> Détail technique : [12-sso-federation.md](../foundation/12-sso-federation.md). Gate : [ROADMAP §Phase 1](../ROADMAP.md).

| Élément | Détail |
| --- | --- |
| Admin UI | `admin/identity-providers` — configurer Azure AD / Google (Enterprise) |
| BFF | `server/api/auth/oidc/*` — proxy authorize/callback |
| Liaison | Rattachement compte `XXX_nom` existant ou création JIT (si sièges disponibles) |

- [ ] IdP configurable par tenant Administrateur.
- [ ] Liaison JIT testée ; refus si plafond sièges (module 14).

## 12. Definition of Done

- [x] Hiérarchie org CRUD complète et testée.
- [x] Auth login/refresh/logout via cookies httpOnly opérationnelle.
- [x] Matrice RBAC §3.3 chargée et appliquée par middleware.
- [x] Isolation multi-tenant vérifiée par test d'intégration.
- [x] RG-ORG-01/02, RG-SEC-01/02 couvertes par des tests nommés.
- [x] Endpoints documentés dans `api/openapi.yaml`.
- [ ] SSO OIDC : admin IdP + liaison utilisateur (Phase 1, cf. §11bis).
