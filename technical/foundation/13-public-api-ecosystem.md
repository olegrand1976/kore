# 13 — API publique et écosystème

> Fondation transverse. Clés API tierces, webhooks sortants et conventions d'intégration partenaire.
> Complète [05-api-conventions.md](05-api-conventions.md). Phase cible : [ROADMAP §Phase 2](../ROADMAP.md).
> Implémentation module : [17-integrations-hub.md](../modules/17-integrations-hub.md).

## 1. Objectif

Rendre Kore **embeddable** dans l'écosystème client (IT/DSI) : accès programmatique limité et notifications sortantes vers les systèmes tiers.

**Distinction des webhooks** :

| Type | Direction | Auth | Exemple |
| --- | --- | --- | --- |
| Entrant | Externe → Kore | Signature | Stripe (`/webhooks/stripe`), PDP (`/webhooks/pdp`) |
| Sortant | Kore → Externe | HMAC secret abonnement | `cra.submitted`, `leave.approved` |

## 2. Authentification API key

- Header : `X-Api-Key: <key>` (ou `Authorization: Bearer kore_live_...`).
- Clé liée à un `tenant_id`, un **scope** (liste de modules/actions autorisées) et une date d'expiration optionnelle.
- Stockage : hash de la clé en base (`integrations.api_keys`) ; la clé en clair n'est montrée qu'à la création.
- Rate-limiting Redis : clé `kore:{tenant}:apikey:ratelimit:{key_id}`, fenêtre glissante (défaut 1000 req/h, configurable).

### Scopes initiaux (Phase 2)

| Scope | Permissions |
| --- | --- |
| `cra:read` | GET timesheets |
| `cra:write` | PUT weeks, POST submit |
| `leave:read` | GET leave-requests, balances |
| `leave:write` | POST leave-requests |
| `invoices:read` | GET invoices (module 09) |
| `webhooks:manage` | CRUD abonnements webhook |

## 3. Routes partenaires

Préfixe : `/api/v1/integrations/` (miroir lecture/écriture des routes internes selon scope).

| Méthode | Chemin | Scope | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/integrations/api-keys` | Admin JWT | Créer une clé API |
| GET | `/api/v1/integrations/api-keys` | Admin JWT | Lister les clés (masquées) |
| DELETE | `/api/v1/integrations/api-keys/{id}` | Admin JWT | Révoquer une clé |
| POST | `/api/v1/integrations/webhooks` | `webhooks:manage` | Créer abonnement webhook |
| GET | `/api/v1/integrations/webhooks` | `webhooks:manage` | Lister abonnements |
| DELETE | `/api/v1/integrations/webhooks/{id}` | `webhooks:manage` | Supprimer abonnement |

Les routes métier (`/api/v1/timesheets`, etc.) acceptent **JWT ou API key** selon le middleware ([05-api-conventions.md](05-api-conventions.md) §5).

## 4. Webhooks sortants

### Modèle

| Champ | Description |
| --- | --- |
| `url` | Endpoint HTTPS du partenaire |
| `secret` | Secret HMAC pour signature `X-Kore-Signature` |
| `events[]` | Liste d'événements souscrits |
| `enabled` | Actif / suspendu |

### Événements canoniques

| Event | Déclencheur | Module source |
| --- | --- | --- |
| `cra.week_submitted` | Validation prévisionnelle semaine | 02 CRA |
| `cra.validated` | Validation définitive manager | 02 CRA |
| `leave.requested` | Nouvelle demande congé | 03 Congés |
| `leave.approved` | Congé validé | 03 Congés |
| `leave.rejected` | Congé refusé | 03 Congés |
| `invoice.prepared` | Facture préparée (payload EN 16931) | 09 Facturation |
| `invoice.status_changed` | Statut PDP mis à jour | 09 Facturation |

### Livraison

- Payload JSON : `{ "id", "type", "tenant_id", "occurred_at", "data": { ... } }`.
- Signature : `HMAC-SHA256(secret, body)` dans `X-Kore-Signature`.
- Retry : backoff exponentiel (1m, 5m, 30m, 2h, 24h) ; max 5 tentatives ; journal dans `integrations.webhook_deliveries`.
- Idempotence : `event_id` UUID unique par événement.

### Port

```go
// platform/httpx ou module 17
type WebhookDispatcher interface {
    Dispatch(ctx context.Context, evt OutboundEvent) error
}
```

Les modules métier publient via ce port (événement domaine → adapter 17).

## 5. Schéma de données (schéma `integrations`)

| Table | Colonnes clés |
| --- | --- |
| `integrations.api_keys` | `id`, `tenant_id`, `name`, `key_hash`, `prefix`, `scopes[]`, `expires_at`, `revoked_at` |
| `integrations.webhook_subscriptions` | `id`, `tenant_id`, `url`, `secret_hash`, `events[]`, `enabled` |
| `integrations.webhook_deliveries` | `id`, `subscription_id`, `event_id`, `payload`, `status`, `attempts`, `last_attempt_at`, `response_code` |
| `integrations.sync_logs` | `id`, `tenant_id`, `connection_id`, `direction`, `status`, `records_count`, `error` |

## 6. OpenAPI et documentation développeur

- Section `integrations` dans `api/openapi.yaml` : auth API key, liste événements, exemples payload.
- Page publique future : `/developers` (module 15 ou site statique) — hors DoD Phase 2 initiale.
- Sandbox : clés `kore_test_...` avec rate-limit réduit et données fictives (option Phase 2+).

## 7. Tests

- Unitaires : génération signature HMAC ; scope insuffisant → 403 ; clé expirée → 401.
- Unitaires : retry backoff ; idempotence `event_id`.
- Intégration : création clé → appel API scoped → révocation → 401.
- Intégration : dispatch webhook vers serveur mock HTTP.

## 8. Definition of Done (fondation API publique — Phase 2)

- [ ] Création/révocation clé API par admin tenant.
- [ ] Middleware API key intégré au pipeline ([05-api-conventions.md](05-api-conventions.md)).
- [ ] Au moins 3 événements sortants livrés (CRA + congés).
- [ ] Retry et journal de livraison opérationnels.
- [ ] Section OpenAPI `integrations` à jour.
- [ ] Rate-limiting Redis actif sur clés API.
