# cap-conges-manager-assist — Contexte manager congés

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `conges.manager_assist` |
| Module | 03 Congés |
| Vague | 3 — P2 |
| Statut | livré |

## Finalité métier

Fournir un **contexte factuel** au manager lors de la validation congés : soldes demandeur, absences équipe sur la période, chevauchements. Persona : manager / RH validateur.

## Classification IA Act

- **Risque** : haut potentiel (Annexe III §4(b) relation de travail)
- **Annexe III** : §4(b) — aide décision RH
- **Art. 6(3)** : **non haut-risque** si contexte factuel uniquement, **pas de recommandation approve/reject**, manager décide via boutons existants — voir [02-capabilities-registry.md](../02-capabilities-registry.md)
- **Art. 26(7)** : information travailleurs requise avant activation tenant

## Garde-fous

- **Interdit** : suggestion « approuver » / « refuser » ou score candidat
- Contexte **collapsible**, lecture seule
- Décision exclusivement via [`conges/validation.vue`](../../../frontend/pages/conges/validation.vue) — approve/reject manuels
- Explicabilité : `GET /api/v1/ai/explain/{requestId}`

## Entrées / sorties

**Entrée** : `leaveRequestId`  
**Sortie** : `{ balances, teamAbsences[], overlaps[], narrative, requestId }`

## Ancrage code

- [`frontend/pages/conges/validation.vue`](../../../frontend/pages/conges/validation.vue)
- [`frontend/composables/useLeave.ts`](../../../frontend/composables/useLeave.ts)
- [`frontend/composables/usePermissions.ts`](../../../frontend/composables/usePermissions.ts)
- `internal/modules/ai/app/service.go` — `BuildManagerContext`

## API

- BFF : `POST /api/ai/conges/manager-context`
- Go : `POST /api/v1/ai/conges/manager-context`

## DoD

- [x] Routes BFF + handlers Go
- [x] Panneau contexte collapsible + `AppAiBadge`
- [x] Zéro bouton approve/reject IA
- [x] Journalisation `ai_request_log`
- [x] Checklist Art. 26(7) dans gouvernance deployer
