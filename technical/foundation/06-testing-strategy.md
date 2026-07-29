# 06 — Stratégie de tests

> Fondation transverse. Standard de test appliqué à chaque brique. Les tests unitaires sont **partie intégrante** de la spécification de chaque module.

## 1. Pyramide de tests

```mermaid
flowchart TB
  U["Tests unitaires (nombreux, rapides) : domain + app"] --> I["Tests d'intégration : adapters postgres (testcontainers)"]
  I --> C["Tests de contrat API : handlers chi + OpenAPI"]
  C --> E["Tests end-to-end / front : vitest + parcours Nuxt (ciblés)"]
```

Priorité au **socle unitaire** : le `domain` et l'`app` concentrent la logique métier et se testent sans I/O.

## 2. Tests unitaires Go (standard obligatoire)

- **Domaine** : test pur des invariants et règles (aucune dépendance). Ex. transition d'état interdite, calcul UO/facturation.
- **Application (use cases)** : test contre des **mocks des ports outbound** (repositories, gateways). Vérifie l'orchestration et l'application des règles de gestion (RG-xxx) et des critères d'acceptation (spec §8).
- Style **table-driven** :

```go
func TestCRAService_Submit(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*mocks.CRARepository)
        input   SubmitCommand
        wantErr error
    }{
        {name: "soumission valide", /* ... */},
        {name: "deja validee -> conflit", wantErr: domain.ErrCRAAlreadyValidated},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) { /* arrange/act/assert */ })
    }
}
```

- **Mocks** : générés (`mockery`) depuis les ports, ou doublures manuelles simples. Un port = un mock substituable (LSP).
- **Horloge** : port `Clock` injecté pour tester les règles temporelles (jours futurs, dernier lundi du mois) de façon déterministe.

## 3. Tests d'intégration

- **Repositories postgres** testés via **testcontainers-go** (PostgreSQL éphémère) + migrations appliquées.
- Vérifient : mapping row↔entité, contraintes (unicité tenant, append-only ETT), requêtes sqlc réelles.
- **Cache Redis** testé via **testcontainers `redis:7`** : préfixe de clés (`kore:{tenant}:...`), TTL, invalidation ciblée, comportement au miss, dégradation si indisponible (cf. [10-cache-redis.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/10-cache-redis.md)).
- **Stripe** testé via **`stripe-mock`** (conteneur) : création de session Checkout, réception et **idempotence** des webhooks, vérification de signature (cf. [11-payments-stripe.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/11-payments-stripe.md)).
- Marqués `//go:build integration` et exécutés dans un job CI dédié.
- Le cache est testé en **unitaire** avec `InMemoryCache` (sans réseau) ; Stripe en unitaire avec un mock du port `PaymentGateway`.

## 4. Tests de contrat API

- Handlers chi testés via `httptest` : codes HTTP, enveloppe d'erreur, application RBAC (401/403), validation DTO.
- Cohérence avec `api/openapi.yaml`.

## 5. Tests frontend (Nuxt)

- **vitest** + `happy-dom` : composables (logique), utils, helpers BFF purs.
- Tests **BFF** (`frontend/tests/bff/`) : cookie → `Authorization`, parsing session JWT, tokens 2FA partiels.
- **Playwright** (`frontend/e2e/smoke/`) : parcours UI smoke sur stack Docker complète (API + Nuxt + Postgres/Redis).
  - Commandes : `make test-frontend`, `make test-e2e`, `npm run test:e2e` (frontend déjà up).
  - Seed : `ADM_admin` / `Admin123!`.
  - Specs : login, dashboard, CRA, congés, TMA, budget, billing, org admin.

## 6. Seuils et qualité

| Couche | Couverture cible | Nature |
| --- | --- | --- |
| domain | ≥ 50 % gate CI (cible long terme > 90 %) | unitaire pur |
| app (use cases) | > 80 % | unitaire + fakes ports |
| adapters postgres | chemins clés | intégration (`//go:build integration`) |
| handlers http | chemins clés + RBAC | contrat |
| frontend | composables/BFF critiques | unitaire Vitest |
| E2E | parcours MVP smoke | Playwright |

- CI bloquante sur : build, lint, `go test ./...`, gate domain, `make test-integration`, `npm test`, job `e2e` Playwright.
- Chaque **critère d'acceptation** de la spec §8 doit correspondre à au moins un test nommé.

## 7. Definition of Done (fondation testing)

- [x] Standard table-driven + fakes des ports documenté.
- [x] testcontainers en place pour l'intégration DB.
- [x] Port `Clock` prévu pour le déterminisme temporel.
- [x] Seuils de couverture et gates CI définis.
- [x] Playwright smoke MVP + job CI `e2e`.

### DoD pour chaque PR feature

1. Nouvelle règle métier → test domain table-driven.
2. Nouveau use case → test app avec fake ports.
3. Nouveau / alter SQL → test intégration repo (contrainte ou round-trip).
4. Nouveau parcours UI critique → spec Playwright smoke ou extension d'une existante.
5. CI verte (backend + frontend + integration + e2e) avant merge.

### Commandes Makefile

| Cible | Rôle |
| --- | --- |
| `make test` | Unitaires Go |
| `make test-frontend` | Vitest |
| `make test-integration` | Postgres testcontainers |
| `make test-e2e` | Playwright (stack Docker test) |
| `make test-all` | Pyramide complète |
| `make smoke` | Smoke API curl (complémentaire) |
