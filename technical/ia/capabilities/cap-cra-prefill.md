# cap-cra-prefill — Pré-remplissage CRA

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `cra.prefill` |
| Module | 02 CRA |
| Vague | 2 — P1 |
| Statut | livré |

## Finalité métier

Proposer des **lignes de temps** à partir des sources métier (TMA affectées, missions déclarées) pour accélérer la saisie hebdomadaire. Persona : consultant / développeur.

## Classification IA Act

- **Risque** : minimal (aide saisie, pas de validation auto)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous RG-CRA-01

- Lignes `origin=manual` **jamais écrasées** par une suggestion prefill
- Suggestions marquées `origin=prefill` — modifiables avant enregistrement
- Pas de soumission ni validation CRA automatique
- Respect RG-CRA-03 : capacité jour non dépassée

## Entrées / sorties

**Entrée** : `timesheetId`, `week`  
**Sortie** : `{ lines: [{ day, duration, sourceType, sourceId, comment }], requestId }`

## Ancrage code

- [`frontend/pages/cra/[id].vue`](../../../frontend/pages/cra/[id].vue)
- [`frontend/components/cra/TimesheetGrid.vue`](../../../frontend/components/cra/TimesheetGrid.vue)
- [`frontend/components/cra/WeekEditor.vue`](../../../frontend/components/cra/WeekEditor.vue)
- `internal/modules/cra/domain/domain.go` — `LineOrigin`, `OriginPrefill`
- `internal/modules/ai/app/service.go` — `SuggestCraPrefill`

## API

- BFF : `POST /api/ai/cra/prefill-suggest`
- Go : `POST /api/v1/ai/cra/prefill-suggest`

## DoD

- [x] Routes BFF + handlers Go
- [x] Invariant RG-CRA-01 testé (`domain_test.go`)
- [x] Bouton « Pré-remplir » + accept/reject par ligne
- [x] Journalisation `ai_request_log`
