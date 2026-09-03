# CLAUDE.md — Contexte projet Kore

Suite **PSA/ESN** modulaire articulée autour du **CRA pivot** (temps → projet → facturation → RH).
Reprise **fonctionnelle** de l'offre historique **B-Hive** (Bee Software, 2008–2013) : réécriture
**greenfield** — aucun code legacy PHP/Flash/Flex dans le dépôt, les sources B-Hive ne servent que
de référence fonctionnelle.

Monolithe modulaire hexagonal : **API Go** (`internal/`, `cmd/`) + **frontend Nuxt 3 SSR/BFF**
(`frontend/`) + **app Flutter** (`mobile/`). Multi-tenant, SaaS facturé par module (Stripe).

## Stack

| Couche | Techno |
| --- | --- |
| API / métier | Go 1.26 (`go 1.26.0` + `toolchain go1.26.5` dans `go.mod`, seule source de vérité — la CI lit `go-version-file: go.mod`), chi v5, pgx v5 + pgxpool, golang-migrate, sqlc (`sqlc.yaml`) |
| DB | PostgreSQL — **un schéma par module** (`cra`, `tma`, `org`, `authx`, `ai`, …) |
| Cache / sessions | Redis (`redis/go-redis/v9`), préfixe `kore:` |
| Frontend web | Nuxt 3 (Vue 3, Nitro), Pinia, `@nuxtjs/i18n` (fr défaut, en) |
| Mobile | Flutter 3.x, OIDC PKCE |
| Paiements | Stripe (`stripe-go`), `stripe-mock` en test |
| Cloud | GCP : Cloud Run, Cloud SQL, Memorystore, Secret Manager, Artifact Registry, Cloud Build |
| Contrat API | OpenAPI 3.1 — `internal/app/openapi.yaml` (embarqué, test de contrat `internal/app/openapi_contract_test.go`) |

## Commandes

```bash
make up                # stack Docker complète (infra + migrate + api + frontend)
make up-infra          # db, redis, mailhog, stripe-mock seuls (dev Go/Nuxt hors conteneur)
make front             # rebuild du seul service frontend
make migrate           # migrations (conteneur one-shot)
make seed / seed-reset # jeu de données demo (idempotent / réinitialisé)
make ready / smoke     # /health + /ready / smoke test API complet
make logs / ps / down

make test              # unitaires Go
make test-frontend     # Vitest
make test-integration  # Postgres via testcontainers (tag `integration`)
make test-e2e          # Playwright smoke (stack Docker test)
make test-all          # pyramide complète
make lint              # golangci-lint
make sqlc              # sqlc generate
```

Ports locaux (surchargeables dans `.env`) : frontend **3001**, API **8081**, Postgres **5434**,
Redis **6381**, MailHog UI **8025**, stripe-mock **12111**.

Comptes seed (`internal/seed/constants.go`) : `ADM_admin` / `Admin123!`, `MGR_manager` /
`Manager123!`, `COL_collab` / `Collab123!`, `PRE_presta`, `CLI_contact`, `COM_commercial`.

## Architecture backend

```
cmd/kore-api/main.go              # entrée
internal/app/app.go               # câblage complet (DI manuelle), montage des routes /api/v1
internal/platform/                # authx, cache, config, cryptox, db, httpx, logging, pdf, uploads
internal/modules/<module>/
  domain/                         # entités, règles, erreurs sentinelles (aucune dépendance externe)
  ports/                          # interfaces inbound (<X>Service) et outbound (<X>Repository/Gateway)
  app/                            # use cases (implémentent les ports inbound)
  adapters/http/                  # handlers chi + RegisterRoutes(...)
  adapters/postgres/              # implémentation des repositories
  adapters/<autre-module>/        # adaptateur de consommation d'un module voisin (ex. cra/adapters/org)
  migrations/                     # NNNN_description.up.sql / .down.sql (schéma du module)
db/migrations/                    # migrations transverses
pkg/kernel/                       # types partagés (TenantID, …)
```

18 modules : `org`, `workflow`, `cra`, `conges`, `budget`, `tma`, `support`, `maintenance`,
`ssii`, `invoicing`, `ett`, `notifications`, `reporting`, `admin`, `billing`, `integrations`,
`publicsite`, `ai`.

**Règle de dépendance** : `domain` ne dépend de rien ; `app` ne connaît que `domain` + `ports` ;
un module n'importe **jamais** l'`app` d'un autre module — il passe par un port + un adaptateur
dédié dans `adapters/<module-voisin>/`.

Un nouveau module se branche dans [internal/app/app.go](internal/app/app.go) : repo postgres →
service app → `<module>http.RegisterRoutes(r, svc, tokenIssuer, authorizer, billingService)`.

## Conventions API

- REST versionné `/api/v1`, ressources plurielles kebab-case, actions métier en sous-ressource
  (`POST /timesheets/{id}/submit`).
- Enveloppe d'erreur uniforme `{ "error": { code, message, details } }` via
  `httpx.WriteError(w, status, code, msg)` ; succès via `httpx.WriteData`.
- Codes notables : `402` module non souscrit, `403` RBAC, `409` conflit d'état, `422` règle métier,
  `429` rate-limit routes publiques.
- Pagination `?page=&page_size=` (défaut 50, max 200), tri `?sort=&order=`, dates ISO 8601 UTC.
- Pipeline d'auth, dans cet ordre : signature webhook → routes publiques (`/api/v1/public/*`,
  rate-limit IP) → API key (`X-Api-Key`, sous `/api/v1/open`) → JWT → **entitlement** (module
  souscrit) → **RBAC** (profil × module × action).
- Middlewares : `httpx.AuthStack(tokens, entitlements)`, `httpx.PublicAPIStack(...)`,
  `authorizer.Can(ctx, "<module>", authx.Action…)` en tête de handler.
- Profils socle RBAC : Utilisateur, Collaborateur, Chef d'équipe, Responsable de service,
  Commercial, Support, Administrateur, Chef utilisateur, Client externe, Sous-traitant.
  Actions : L (lecture), E (écriture), V (validation).

## Frontend Nuxt

- **Le client n'appelle jamais l'API Go directement** : tout passe par le BFF Nitro
  `frontend/server/api/**`, qui relaie vers `${apiBase()}/api/v1/...` avec `apiAuthHeaders(event)`
  (lit le cookie httpOnly → header `Authorization: Bearer`). Voir
  [frontend/server/utils/auth.ts](frontend/server/utils/auth.ts).
- Auth : cookies httpOnly `kore_access_token` / `kore_refresh_token`, session via
  `/api/auth/session`, refresh transparent sur 401 dans
  [useApiFetch.ts](frontend/composables/useApiFetch.ts) (`useRequestFetch` pour propager le cookie en SSR).
- Middlewares : `auth.global.ts` (allowlist de routes publiques), `admin.ts`
  (`profile === 'Administrateur'`), `cra-gate.global.ts`, `platform.ts`.
- Layouts : `public.vue` (site vitrine) / `default.vue` (app). Mobile : `MobileDrawer.vue`,
  `AppBottomNav.vue`.
- Composants préfixés par domaine (`components/cra/`, `budget/`, `tma/`, `ui/`, `public/`, …),
  `pathPrefix: false` → nom de fichier = nom de composant.
- **Charte** : uniquement les tokens `--kore-*` de
  [frontend/assets/css/tokens.css](frontend/assets/css/tokens.css), jamais de couleur ad hoc.
  Thème dark par défaut (`useTheme.ts`). Icônes Material Symbols via `AppIcon.vue`.
- **i18n obligatoire** : zéro string FR hardcodée dans les templates → `frontend/locales/{fr,en}.json`.

## Tests

| Niveau | Emplacement / commande |
| --- | --- |
| Domain / app Go | `*_test.go` à côté du code, fakes pour les ports — `make test` (gate de couverture `scripts/check-coverage.sh`) |
| Intégration Go | `*_integration_test.go` (tag `integration`, testcontainers) — `make test-integration` |
| Vitest | `frontend/tests/` (dont `tests/bff/`) — `make test-frontend` |
| Playwright | `frontend/e2e/smoke/` — `make test-e2e` |
| Smoke API | `scripts/smoke-test.sh` — `make smoke` |

CI ([.github/workflows/ci.yml](.github/workflows/ci.yml)) : jobs `backend`, `integration`,
`frontend`, `mobile`, `smoke`. Stratégie détaillée :
[technical/foundation/06-testing-strategy.md](technical/foundation/06-testing-strategy.md).

## Règles projet à respecter

1. **Migration jamais modifiée après merge** → toute évolution = nouvelle migration numérotée.
2. Toute migration met à jour [documentation/SCHEMA_DB.md](documentation/SCHEMA_DB.md) **dans la
   même PR**, et ajoute un `*_integration_test.go` si contrainte/round-trip à vérifier.
2bis. Toute évolution RBAC / menus met à jour l'Aide in-app (`frontend/pages/aide/`, i18n `help.*`)
   et [documentation/GUIDE_ACCES_UTILISATEURS.md](documentation/GUIDE_ACCES_UTILISATEURS.md)
   **dans la même PR** (miroir `DefaultPermissions` ↔ `frontend/utils/rbac.ts`).
2ter. Toute **nouvelle fonctionnalité** met à jour le **flux de tests** dans la **même PR**
   (domain/app Go, intégration si SQL, Vitest BFF/composables, Playwright smoke si parcours
   critique, `scripts/smoke-test.sh` si surface API) — cf. règle `feature-tests-sync` et
   [technical/foundation/06-testing-strategy.md](technical/foundation/06-testing-strategy.md).
3. Nouvel endpoint → `internal/app/openapi.yaml` mis à jour (test de contrat sinon rouge).
4. Aucun secret committé ; toute clé nouvelle documentée dans `.env.example` (prod = Secret Manager).
5. Avant PR UI : responsive ≤768px (drawer + bottom nav, CTAs pleine largeur), i18n, tokens charte,
   `npm run build` + `go build ./...`, spec Playwright si parcours critique.
6. Aucun token/secret exposé au client : tout appel tiers (GitHub API des release notes, Stripe,
   Gemini) part du BFF ou du backend.
7. Erreurs métier = sentinelles dans `domain`, enveloppées `%w`, traduites en code HTTP dans
   `adapters/http` — pas de `panic` en flux nominal.

## Déploiement

- Branche `staging` → déploiement GCP automatique
  ([.github/workflows/deploy-gcp-staging.yml](.github/workflows/deploy-gcp-staging.yml)),
  postdeploy `make gcp-postdeploy-staging` (seed reset + smoke). Projet GCP : `premedica-prod-2025`.
- Cibles utiles : `make gcp-setup`, `gcp-deploy`, `gcp-deploy-jobs`, `gcp-postdeploy`, `gcp-smoke`,
  `gcp-domain` (kore.ll-it-sc.be), `setup-oidc-google`.
- Versioning : tags git SemVer uniquement, décidés en CI par évaluation IA
  ([scripts/ai-semver-bump.mjs](scripts/ai-semver-bump.mjs)) avec fallback Conventional Commits.
  Pas de bump `package.json`/`go.mod`.
- Wiki GitHub synchronisé au deploy ([scripts/sync-github-wiki.sh](scripts/sync-github-wiki.sh)).

## Où chercher la spec

| Sujet | Fichier |
| --- | --- |
| Fonctionnel complet (modules, processus, RG) | [documentation/SPECIFICATION_FONCTIONNELLE.md](documentation/SPECIFICATION_FONCTIONNELLE.md) |
| Accès par profil (Aide in-app) | [documentation/GUIDE_ACCES_UTILISATEURS.md](documentation/GUIDE_ACCES_UTILISATEURS.md) — UI `/aide` |
| Schéma DB | [documentation/SCHEMA_DB.md](documentation/SCHEMA_DB.md) |
| Charte graphique | [documentation/CHARTE_GRAPHIQUE.md](documentation/CHARTE_GRAPHIQUE.md) |
| Index technique + ordre de construction | [technical/README.md](technical/README.md) |
| Fondations transverses (archi, auth, API, tests, GCP, SSO…) | [technical/foundation/](technical/foundation/) |
| Fiche par module métier | [technical/modules/](technical/modules/) (00→17) |
| IA (conformité IA Act, registre capabilities) | [technical/ia/](technical/ia/) |
| Phases, gates, dette | [technical/ROADMAP.md](technical/ROADMAP.md) |
| Conventions agents / sub-agents Cursor | [.cursor/AGENTS.md](.cursor/AGENTS.md) |

## Langue et commits

- Identifiants, types et commentaires de code : **anglais**.
- UI, messages utilisateur, documentation : **français**.
- Commits : **Conventional Commits avec sujet en français** —
  `feat(cra): lignes dupliquées et affectations TMA`, `fix(ci): …`, `docs: …`.

## Pièges connus

- **`gofmt` bloque régulièrement le déploiement staging** (`golangci-lint` en CI) : lancer
  `gofmt -l .` / `make lint` avant de pousser sur `staging`.
- Le cache Redis peut masquer un `seed-reset` : invalider ou redémarrer Redis si les données seed
  semblent périmées.
- `internal/app/openapi.yaml` incomplet = CI rouge (test de contrat), y compris pour les webhooks.
- Codes de workflow contenant un point : attention aux routes `/admin/workflows/[code]`.
