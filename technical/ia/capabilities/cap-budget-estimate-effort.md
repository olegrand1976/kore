# cap-budget-estimate-effort — Estimation effort budget

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `budget.estimate` |
| Module | 04 Budget |
| Vague | 2 — P1 |
| Statut | livré |

## Finalité métier

Proposer une **estimation** (jours + UO) à partir du sujet et du contexte demande TMA, pour accélérer la saisie manager budget. Persona : manager budget / chef de projet.

## Classification IA Act

- **Risque** : minimal (aide chiffrage, devis prime toujours)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- Suggestion **non enregistrée** tant que le manager ne soumet pas le formulaire estimation
- Le **devis prime** sur l'estimation (RG métier brique 04)
- Pas de déclenchement recompute automatique

## Entrées / sorties

**Entrée** : `budgetId`, `demandId`, `subject` (optionnel)  
**Sortie** : `{ effortDays, effortUO, rationale, requestId }`

## Ancrage code

- [`frontend/pages/budget/[id].vue`](../../../frontend/pages/budget/[id].vue)
- [`frontend/components/budget/BudgetTripleGauge.vue`](../../../frontend/components/budget/BudgetTripleGauge.vue)
- [`frontend/composables/useBudget.ts`](../../../frontend/composables/useBudget.ts)
- `internal/modules/ai/app/service.go` — `SuggestBudgetEstimate`

## API

- BFF : `POST /api/ai/budget/estimate-effort`
- Go : `POST /api/v1/ai/budget/estimate-effort`

## DoD

- [x] Routes BFF + handlers Go
- [x] Bouton « Estimer » + pré-remplissage formulaire
- [x] Journalisation `ai_request_log`
- [x] Badge + disclaimer
