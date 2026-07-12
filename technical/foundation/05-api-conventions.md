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

- En-tête `Idempotency-Key` accepté sur les POST d'action sensibles (préparation facture).
- Dates en ISO 8601 UTC.
- Toutes les routes (sauf `auth/*`, `webhooks/*`, `public/*`, `health`) passent par les middleware : auth JWT -> résolution tenant -> **entitlement (module souscrit, cf. 04/11)** -> RBAC -> handler. Les webhooks (`webhooks/stripe`, `webhooks/pdp`) sont authentifiés par **signature**, hors JWT.
- Les routes **publiques** (`/api/v1/public/*`, cf. [module 15](/home/olivier/ll-it-sc/projets/kore/technical/modules/15-site-vitrine-booking.md)) sont **non authentifiées** mais protégées par **rate-limiting Redis** + anti-spam ; elles ne portent ni JWT ni tenant applicatif.
- `GET /health` (liveness) et `GET /ready` (readiness DB) non authentifiés.

## 6. Documentation

- Chaque fiche module liste ses endpoints (méthode, chemin, DTO, permission RBAC requise, codes d'erreur).
- L'`openapi.yaml` est mis à jour à chaque brique et sert de base aux tests de contrat et à la génération des types côté Nuxt.

## 7. Definition of Done (fondation API)

- [ ] Enveloppe d'erreur et codes HTTP actés.
- [ ] Conventions pagination/tri/filtre définies.
- [ ] Pipeline de middleware ordonné (auth -> tenant -> rbac).
- [ ] Squelette `api/openapi.yaml` initialisé.
