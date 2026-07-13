# cap-tma-classify-duplicate — Classification et doublons TMA

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `tma.classify`, `tma.similar` |
| Module | 05 TMA |
| Vague | 1 — P0 |
| Statut | livré |

## Finalité métier

À la création d'une demande : proposer un **type** (incident, évolution, régression…) et alerter sur des **demandes similaires** déjà résolues sur la même application. Persona : développeur TMA / chef utilisateur.

## Classification IA Act

- **Risque** : minimal (aide à la saisie, pas de décision workflow)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous RG-TMA-01

- Ne modifie pas le statut ni la gate chef utilisateur
- Classification et doublons **informatifs** — l'utilisateur confirme le sujet et le type avant soumission
- Similarité : seuil configurable, jamais de fusion ou clôture auto

## Entrées / sorties

**`tma.classify`** — entrée : `subject`, `applicationId` — sortie : `{ type, confidence, requestId }`  
**`tma.similar`** — entrée : `subject`, `applicationId`, `limit` — sortie : `{ items: [{ demandId, subject, score }], requestId }`

## Ancrage code

- [`frontend/components/tma/DemandForm.vue`](../../../frontend/components/tma/DemandForm.vue)
- [`frontend/pages/tma/index.vue`](../../../frontend/pages/tma/index.vue)
- `internal/modules/ai/app/service.go` — `ClassifyDemand`, `FindSimilarDemands`

## API

- BFF : `POST /api/ai/tma/classify`, `GET /api/ai/tma/similar`
- Go : `POST /api/v1/ai/tma/classify`, `GET /api/v1/ai/tma/similar`

## DoD

- [x] Routes BFF + handlers Go
- [x] Stub : mots-clés + recherche textuelle LIKE
- [x] Journalisation `ai_request_log`
- [x] Bandeau doublons + badge type dans `DemandForm`
