# cap-publicsite-chatbot — Chatbot site public

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `publicsite.chatbot` |
| Module | 15 Public |
| Vague | 3 — P2 |
| Statut | livré |

## Finalité métier

Répondre aux questions **visiteurs** sur Kore (offre PSA, modules, tarifs, démo) via un assistant conversationnel sur le site public. Persona : prospect / visiteur non authentifié.

## Classification IA Act

- **Risque** : limité — **Art. 50** (interaction avec système IA)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- **Sans JWT** — rate-limit BFF 20 req/min/IP
- Pas d'accès données tenant ni PII
- Disclosure Art. 50 : `AppAiBadge variant="assistant"` + notice chatbot
- Pas de collecte lead sans consentement explicite (formulaire séparé)

## Entrées / sorties

**Entrée** : `message`, `sessionId` (anonyme)  
**Sortie** : `{ reply, sessionId, requestId }`

## Ancrage code

- [`frontend/pages/index.vue`](../../../frontend/pages/index.vue)
- [`frontend/pages/reserver/index.vue`](../../../frontend/pages/reserver/index.vue)
- [`frontend/pages/tarifs/index.vue`](../../../frontend/pages/tarifs/index.vue)
- `internal/modules/ai/app/service.go` — `PublicChat`

## API

- BFF : `POST /api/ai/public/chat` (public, sans auth)
- Go : `POST /api/v1/ai/public/chat`

## DoD

- [x] Routes BFF + handlers Go
- [x] Widget chat + badge assistant + disclosure Art. 50
- [x] Rate-limit documenté
- [x] Journalisation `ai_request_log` (session anonyme)
