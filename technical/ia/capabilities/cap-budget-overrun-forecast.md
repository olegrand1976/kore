# cap-budget-overrun-forecast — Prévision dépassement budget

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `budget.overrun_forecast` |
| Module | 04 Budget |
| Vague | 2 — P2 |
| Statut | livré |

## Finalité métier

Projeter un **risque de dépassement** (jours/UO/montant) à horizon fin de période, à partir de la consommation courante et du rythme CRA/TMA. Persona : manager budget / direction.

## Classification IA Act

- **Risque** : minimal (projection indicative, pas décision financière auto)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- Entrées **agrégées factuelles** (consommation, planifié, tendance linéaire)
- Alerte informative — pas de blocage consommation ni facturation auto
- Mention explicite « estimation indicative »

## Entrées / sorties

**Entrée** : `budgetId`  
**Sortie** : `{ forecastDays, forecastUO, overrunRisk, narrative, requestId }`

## Ancrage code

- [`frontend/pages/budget/index.vue`](../../../frontend/pages/budget/index.vue)
- [`frontend/pages/budget/[id].vue`](../../../frontend/pages/budget/[id].vue)
- [`frontend/components/budget/BudgetTripleGauge.vue`](../../../frontend/components/budget/BudgetTripleGauge.vue)
- `internal/modules/ai/app/service.go` — `ForecastBudgetOverrun` (à créer)

## API

- BFF : `GET /api/ai/budget/overrun-forecast`
- Go : `GET /api/v1/ai/budget/overrun-forecast`

## DoD

- [ ] Route BFF + handler Go
- [ ] Journalisation `ai_request_log`
- [ ] Indicateur prévision sur liste et détail budget
- [ ] i18n fr/en
