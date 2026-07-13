# cap-conges-date-suggest — Suggestion dates congés

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `conges.date_suggest` |
| Module | 03 Congés |
| Vague | 3 — P2 |
| Statut | planifié |

## Finalité métier

Proposer des **créneaux de dates** cohérents (soldes CP, jours fériés, chevauchements équipe) lors de la création d'une demande. Persona : salarié.

## Classification IA Act

- **Risque** : limité (aide planification, pas de décision manager)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- Suggestions **indicatives** — dates modifiables avant soumission
- Pas d'approbation ni rejet automatique
- Données équipe : agrégats d'absence uniquement (pas de classement individuel)

## Entrées / sorties

**Entrée** : `type`, `durationDays`, `teamId` (optionnel)  
**Sortie** : `{ suggestions: [{ from, to, rationale }], requestId }`

## Ancrage code

- [`frontend/pages/conges/index.vue`](../../../frontend/pages/conges/index.vue)
- [`frontend/pages/conges/soldes.vue`](../../../frontend/pages/conges/soldes.vue)
- [`frontend/composables/useLeave.ts`](../../../frontend/composables/useLeave.ts)
- `internal/modules/ai/app/service.go` — `SuggestLeaveDates` (à créer)

## API

- BFF : `POST /api/ai/conges/date-suggest`
- Go : `POST /api/v1/ai/conges/date-suggest`

## DoD

- [ ] Route BFF + handler Go
- [ ] Journalisation `ai_request_log`
- [ ] Chips suggestions dans formulaire demande
- [ ] i18n fr/en
