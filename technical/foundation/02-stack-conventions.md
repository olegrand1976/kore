# 02 — Stack technique et conventions

> Fondation transverse. Définit les briques techniques et les conventions communes à tous les modules.

## 0. Contexte : legacy B-Hive vs stack Kore

| | B-Hive historique (2008–2013) | Kore (dépôt actuel) |
| --- | --- | --- |
| Rôle | Référence **fonctionnelle** (42 docs sources) | Application **greenfield** en production de développement |
| Stack | PHP / Flash / Flex | **Go + PostgreSQL + Redis + Nuxt 3** |
| Code dans ce dépôt | Aucun | `cmd/kore-api/`, `internal/modules/`, `frontend/` |

La modernisation n'est **pas** un masquage ou une surcouche de l'ancienne interface Flex/Flash : c'est une réécriture complète. Les sources B-Hive alimentent la spec fonctionnelle ; la stack ci-dessous est la **décision technique actée**.

## 1. Stack retenue

| Couche | Technologie | Rôle |
| --- | --- | --- |
| Langage backend | **Go** (>= 1.23) | API et logique métier |
| Routeur HTTP | **chi** (`go-chi/chi/v5`) | Routing, middleware, léger, proche stdlib |
| Accès données | **sqlc** | Génération de code type-safe depuis SQL |
| Driver PostgreSQL | **pgx** (`jackc/pgx/v5`) + `pgxpool` | Pool de connexions performant |
| Migrations | **golang-migrate** | Migrations SQL versionnées par schéma |
| Base de données | **PostgreSQL** (>= 16) — **Cloud SQL** en prod | Persistance, un schéma par module |
| Cache | **Redis** (`redis/go-redis/v9`) — **Memorystore partagé** en prod | Cache-aside, sessions/révocation ([10-cache-redis.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/10-cache-redis.md)) |
| Paiements | **Stripe** (`stripe-go`) | Abonnements SaaS ([11-payments-stripe.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/11-payments-stripe.md)) |
| Cloud | **GCP** : Cloud Run, Cloud SQL, Memorystore, Secret Manager, Artifact Registry, Cloud Build | Déploiement ([09-gcp-infrastructure.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/09-gcp-infrastructure.md)) |
| Frontend web | **Nuxt 3** (Vue 3, Nitro) | SSR + BFF (server routes) |
| Mobile | **Flutter 3.x** (Dart) | App iOS/Android ([14-flutter-mobile-client.md](14-flutter-mobile-client.md)) |
| État frontend web | **Pinia** | Store par module |
| Icônes | **Google Fonts — Material Symbols** | Iconographie ([08-frontend-nuxt.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/08-frontend-nuxt.md)) |
| Conteneurisation | **Docker** + Docker Compose | Dev et tests locaux |
| Tests Go | `testing` + `testify` + `testcontainers-go` | Unitaire + intégration (PostgreSQL, Redis) |
| Tests front web | `vitest` + `@vue/test-utils` | Composants et composables Nuxt |
| Tests mobile | `flutter_test` + `mockito` | Widgets et repositories Flutter |
| Mocks Go | `mockery` (ou mocks manuels) | Doublure des ports |
| Paiements (test) | **`stripe-mock`** | Tests d'intégration Stripe |
| Contrat API | **OpenAPI 3.1** | Source de vérité des endpoints |

## 2. Justification des choix structurants

- **chi** : idiomatique, compatible `net/http`, middleware composables, aucun framework lourd → testabilité maximale.
- **sqlc** : le SQL reste explicite et auditable (important pour la facturation et l'inaltérabilité ETT), tout en offrant des signatures Go typées. Pas de magie ORM cachée.
- **pgx** : support natif des types PostgreSQL, performances, `pgxpool` pour la concurrence.
- **golang-migrate** : migrations `up/down` versionnées, exécutables au boot ou via CLI, compatibles CI.

## 3. Conventions de code Go

- Formatage : `gofmt` + `goimports` obligatoires. Lint : `golangci-lint` (govet, staticcheck, errcheck, revive, depguard).
- Nommage packages : court, minuscule, sans underscore (`cra`, `tma`, `authx`).
- Interfaces (ports) : nommées par capacité (`CRAReader`, `InvoicePreparer`), suffixe `-er` quand pertinent (ISP).
- Erreurs : erreurs sentinelles métier dans `domain` (`var ErrCRAAlreadyValidated = errors.New(...)`), enveloppées avec `%w`. Pas de `panic` en flux nominal.
- Contexte : `context.Context` en premier argument de toute méthode de port I/O.
- Pas d'import inline : tous les imports en tête de fichier.
- Injection : dépendances passées par constructeur (`NewCRAService(repo CRARepository, clock Clock) *CRAService`).

## 4. Conventions de nommage transverses

| Élément | Convention | Exemple |
| --- | --- | --- |
| Schéma DB | nom du module | `cra`, `tma`, `org` |
| Table | pluriel snake_case | `cra.timesheets` |
| Migration | `NNNN_description.up.sql` / `.down.sql` | `0001_create_timesheets.up.sql` |
| Endpoint REST | kebab-case, ressource plurielle | `/api/v1/timesheets` |
| DTO | suffixe `Request` / `Response` | `CreateTimesheetRequest` |
| Port inbound | `<Domaine>Service` | `CRAService` |
| Port outbound | `<Domaine>Repository` / `<X>Gateway` | `CRARepository`, `PDPGateway` |

## 5. Configuration

- Chargement via variables d'environnement (12-factor), typées dans `platform/config`.
- Aucune valeur secrète en dur ni committée. `.env.example` documenté (cf. 07-docker-devops.md). En prod : **Secret Manager** (cf. 09-gcp-infrastructure.md).
- Clés principales : `DATABASE_URL`, `HTTP_ADDR`, `JWT_SIGNING_KEY`, `JWT_TTL`, `TENANT_MODE`, `LOG_LEVEL`.
- Cache : `REDIS_ADDR`, `REDIS_AUTH`, `REDIS_KEY_PREFIX`, `REDIS_TLS`, `CACHE_DEFAULT_TTL`.
- Paiements : `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PUBLISHABLE_KEY`, `STRIPE_PRICE_<MODULE>`, `BILLING_TRIAL_DAYS`.

## 6. Versionnage et compatibilité

- API versionnée par préfixe d'URL (`/api/v1`).
- Migrations jamais modifiées après merge : toute évolution = nouvelle migration.
- Changement cassant d'un port = revue d'impact sur les modules consommateurs.

## 7. Definition of Done (fondation stack)

- [ ] Versions figées (Go, PostgreSQL, Nuxt) dans la doc et les Dockerfile.
- [ ] `golangci-lint` configuré et vert.
- [ ] Conventions de nommage appliquées dans les fiches modules.
