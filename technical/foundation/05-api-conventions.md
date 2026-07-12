# 05 — Conventions d'API REST

> Fondation transverse. Contrat d'API commun à toutes les briques (adapters HTTP chi).

## 1. Principes

- **REST** orienté ressources, JSON, versionné par préfixe `/api/v1`.
- Source de vérité : **OpenAPI 3.1** (`api/openapi.yaml`), agrégeant les endpoints de chaque module.
- Chaque module expose son routeur chi monté sous son préfixe (`/api/v1/timesheets`, `/api/v1/demands`, ...).

## 2. Verbes et ressources

| Verbe | Usage | Idempotent |
| --- | --- | --- |
| GET | Lecture (collection/élément) | Oui |
| POST | Création / action métier | Non |
| PUT | Remplacement complet | Oui |
| PATCH | Mise à jour partielle | Non |
| DELETE | Suppression / archivage | Oui |

Actions métier non-CRUD : sous-ressource explicite (`POST /timesheets/{id}/submit`, `POST /invoices/{id}/prepare`).

## 3. Enveloppe de réponse et erreurs

Réponse succès : corps = ressource ou `{ "data": ..., "meta": ... }` pour les collections.

Erreur (format uniforme) :

```json
{
  "error": {
    "code": "CRA_ALREADY_VALIDATED",
    "message": "Le CRA est déjà validé pour cette semaine.",
    "details": [{ "field": "week", "issue": "locked" }]
  }
}
```

| HTTP | Usage |
| --- | --- |
| 200/201/204 | Succès |
| 400 | Validation entrée |
| 401 | Non authentifié |
| 402 | Paiement requis (module non souscrit, cf. module 14) |
| 403 | Non autorisé (RBAC) |
| 404 | Ressource absente (ou hors tenant) |
| 409 | Conflit métier (état/workflow) |
| 410 | Ressource expirée (ex. créneau de réservation) |
| 422 | Règle de gestion violée |
| 429 | Trop de requêtes (rate-limit routes publiques) |
| 500 | Erreur inattendue |

Mapping : les erreurs sentinelles `domain` sont traduites en `code`+HTTP par un helper `platform/httpx`.

## 4. Pagination, tri, filtrage

- Pagination : `?page=1&page_size=50` (défaut 50, max 200), `meta.total` renvoyé.
- Tri : `?sort=created_at&order=desc`.
- Filtrage : paramètres explicites (`?status=submitted&application_id=...`).

## 5. Conventions transverses

### Pipeline d'authentification (ordre d'évaluation)

1. **Signature webhook** (`webhooks/stripe`, `webhooks/pdp`) — hors JWT/API key.
2. **Routes publiques** (`/api/v1/public/*`) — rate-limit IP uniquement.
3. **API key** (`X-Api-Key`) — scope + rate-limit tenant ([13-public-api-ecosystem.md](13-public-api-ecosystem.md)).
4. **JWT** (cookie via BFF ou `Authorization: Bearer`) — refresh si expiré.
5. **Entitlement** (module souscrit, module 14).
6. **RBAC** (profil × module × action).

Erreurs auth : `401 INVALID_CREDENTIALS`, `401 INVALID_API_KEY`, `401 TOKEN_EXPIRED`, `403 INSUFFICIENT_SCOPE`.

### Autres conventions

- En-tête `Idempotency-Key` accepté sur les POST d'action sensibles (préparation facture).
- Dates en ISO 8601 UTC.
- Toutes les routes métier passent par le pipeline ci-dessus (sauf `auth/*` login public, `health`, `ready`).
- Les routes **publiques** (`/api/v1/public/*`, cf. [module 15](../modules/15-site-vitrine-booking.md)) sont **non authentifiées** mais protégées par **rate-limiting Redis** + anti-spam ; elles ne portent ni JWT ni tenant applicatif.
- `GET /health` (liveness) et `GET /ready` (readiness DB) non authentifiés.

## 6. Documentation

- Chaque fiche module liste ses endpoints (méthode, chemin, DTO, permission RBAC requise, codes d'erreur).
- L'`openapi.yaml` est mis à jour à chaque brique et sert de base aux tests de contrat et à la génération des types côté Nuxt.

## 7. Definition of Done (fondation API)

- [x] Enveloppe d'erreur et codes HTTP actés.
- [x] Conventions pagination/tri/filtre définies.
- [x] Pipeline de middleware ordonné (auth -> tenant -> rbac).
- [x] Squelette `api/openapi.yaml` initialisé.
- [ ] Middleware API key et webhooks sortants (cf. [13-public-api-ecosystem.md](13-public-api-ecosystem.md), [ROADMAP Phase 2](../ROADMAP.md)).
