# cap-workflow-explain — Explication workflow

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `workflow.explain` |
| Module | 01 Workflow |
| Vague | 3 — P2 |
| Statut | livré |

## Finalité métier

Expliquer en langage clair **l'état courant**, les **actions disponibles** et les **prérequis** d'une instance workflow (TMA, congés, support). Persona : utilisateur métier non expert workflow.

## Classification IA Act

- **Risque** : minimal (aide compréhension, pas d'exécution)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- Basé sur l'**état machine** et les gardes documentées — pas d'invention d'actions
- N'exécute **aucune transition** — les boutons `WorkflowActions` restent la seule voie
- Aligné RG-TMA-01 pour la gate chef utilisateur

## Entrées / sorties

**Entrée** : `instanceId`  
**Sortie** : `{ stateLabel, explanation, availableActions[], blockers[], requestId }`

## Ancrage code

- [`frontend/components/tma/WorkflowActions.vue`](../../../frontend/components/tma/WorkflowActions.vue)
- [`frontend/pages/tma/[id].vue`](../../../frontend/pages/tma/[id].vue)
- [`frontend/composables/useWorkflow.ts`](../../../frontend/composables/useWorkflow.ts)
- `internal/modules/ai/app/service.go` — `ExplainWorkflowInstance`

## API

- BFF : `GET /api/ai/workflow/explain`
- Go : `GET /api/v1/ai/workflow/explain`

## DoD

- [x] Routes BFF + handlers Go
- [x] Encart « Pourquoi ces actions ? » sur détail TMA
- [x] Journalisation `ai_request_log`
- [x] Lien explicabilité `GET /api/v1/ai/explain/{requestId}`
