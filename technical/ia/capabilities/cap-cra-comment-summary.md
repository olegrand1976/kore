# cap-cra-comment-summary — Résumé commentaires CRA

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `cra.comment_summary` |
| Module | 02 CRA |
| Vague | 2 — P2 |
| Statut | livré |

## Finalité métier

Condenser les **commentaires** d'une semaine CRA en un paragraphe lisible pour le manager (revue mensuelle, export PDF). Persona : manager / RH.

## Classification IA Act

- **Risque** : limité (contenu généré Art. 50)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous RG-CRA-01

- Résumé **non persisté** dans le CRA tant que non validé
- Ne modifie pas les commentaires source
- Minimisation PII dans le prompt (pas de données hors périmètre feuille)

## Entrées / sorties

**Entrée** : `timesheetId`, `week`  
**Sortie** : `{ summary, requestId }`

## Ancrage code

- [`frontend/pages/cra/[id].vue`](../../../frontend/pages/cra/[id].vue)
- [`frontend/components/cra/WeekEditor.vue`](../../../frontend/components/cra/WeekEditor.vue)
- `internal/modules/ai/app/service.go` — `SummarizeCraComments` (à créer)

## API

- BFF : `POST /api/ai/cra/comment-summary`
- Go : `POST /api/v1/ai/cra/comment-summary`

## DoD

- [ ] Route BFF + handler Go
- [ ] Journalisation `ai_request_log`
- [ ] Bouton « Résumer » + `AppAiBadge` dans `WeekEditor`
- [ ] i18n fr/en
