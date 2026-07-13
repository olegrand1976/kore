# cap-budget-demand-autocomplete — Autocomplétion demande budget

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `budget.demand_suggest` |
| Module | 04 Budget |
| Vague | 2 — P1 |
| Statut | livré |

## Finalité métier

Suggérer l'**identifiant demande** et le sujet associé lors de la saisie estimation/devis, à partir des demandes TMA ouvertes sur le budget. Persona : manager budget.

## Classification IA Act

- **Risque** : minimal (recherche assistée)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- Suggestions issues des **données tenant** uniquement (pas d'inférence externe)
- Sélection manuelle obligatoire — pas de création estimation auto
- Filtrage par `budgetId` / application liée

## Entrées / sorties

**Entrée** : `budgetId`, `query` (fragment sujet ou id)  
**Sortie** : `{ items: [{ demandId, subject, status }], requestId }`

## Ancrage code

- [`frontend/pages/budget/[id].vue`](../../../frontend/pages/budget/[id].vue)
- [`frontend/composables/useBudget.ts`](../../../frontend/composables/useBudget.ts)
- `internal/modules/ai/app/service.go` — `SuggestBudgetDemand`

## API

- BFF : `GET /api/ai/budget/demand-suggest`
- Go : `GET /api/v1/ai/budget/demand-suggest`

## DoD

- [x] Routes BFF + handlers Go
- [x] Autocomplétion champs `demandId` (estimation + devis)
- [x] Journalisation `ai_request_log`
- [x] Stub : recherche textuelle sur demandes ouvertes
