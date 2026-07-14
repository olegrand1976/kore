# cap-tma-suggest-assignee — Suggestion d'affectation TMA

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `tma.suggest_assignee` |
| Module | 05 TMA |
| Vague | 2 — P1 |
| Statut | livré |

## Finalité métier

Proposer un **assigné candidat** lors de l'action « Affecter », à partir de la charge courante et des compétences déclarées (pas d'historique individuel de performance). Persona : manager TMA.

## Classification IA Act

- **Risque** : limité → potentiel haut (Annexe III §4(b) allocation tâches)
- **Annexe III** : Potentiel §4(b)
- **Art. 6(3)** : **non haut-risque** si suggestion uniquement, basée charge/compétences déclarées, confirmation manuelle via select existant — voir [02-capabilities-registry.md](../02-capabilities-registry.md)

## Garde-fous RG-TMA-01

- **Pas d'assignation automatique** — pré-remplit le select, manager confirme
- Exclut tout scoring de performance individuel ou traits personnels
- Désactivé si tenant IA non activé

## Entrées / sorties

**Entrée** : `demandId`, `applicationId`  
**Sortie** : `{ suggestedUserId, rationale, alternatives[], requestId }`

## Ancrage code

- [`frontend/components/tma/WorkflowActions.vue`](../../../frontend/components/tma/WorkflowActions.vue)
- [`frontend/pages/tma/[id].vue`](../../../frontend/pages/tma/[id].vue)
- `internal/modules/ai/app/service.go` — `SuggestAssignee` (à créer)

## API

- BFF : `POST /api/ai/tma/suggest-assignee`
- Go : `POST /api/v1/ai/tma/suggest-assignee`

## DoD

- [ ] Évaluation Art. 6(3) validée produit + juridique
- [ ] Route BFF + handler Go
- [ ] Journalisation `ai_request_log`
- [ ] Bouton « Suggérer » + `AppAiBadge` dans `WorkflowActions`
- [ ] i18n fr/en
