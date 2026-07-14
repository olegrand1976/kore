# 02 — Registre des capabilities IA

> Matrice centralisée : classification IA Act, vague, statut, évaluation Art. 6(3).
> Référence : [00-ai-act-conformite.md](00-ai-act-conformite.md).

## Matrice capabilities

| Code | Module | Vague | Priorité | Risque IA Act | Annexe III | Statut | Fiche |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `tma.analysis_draft` | 05 TMA | 1 | P0 | minimal/limité | Non | livré | [cap-tma-analysis-draft.md](capabilities/cap-tma-analysis-draft.md) |
| `tma.classify` | 05 TMA | 1 | P0 | minimal | Non | livré | [cap-tma-classify-duplicate.md](capabilities/cap-tma-classify-duplicate.md) |
| `tma.similar` | 05 TMA | 1 | P0 | minimal | Non | livré | [cap-tma-classify-duplicate.md](capabilities/cap-tma-classify-duplicate.md) |
| `tma.suggest_assignee` | 05 TMA | 2 | P1 | limité → haut | Potentiel §4(b) | livré | [cap-tma-suggest-assignee.md](capabilities/cap-tma-suggest-assignee.md) |
| `tma.executive_summary` | 05 TMA | 2 | P2 | limité | Non | livré | [cap-tma-executive-summary.md](capabilities/cap-tma-executive-summary.md) |
| `cra.prefill` | 02 CRA | 2 | P1 | minimal | Non | livré | [cap-cra-prefill.md](capabilities/cap-cra-prefill.md) |
| `cra.anomalies` | 02 CRA | 2 | P1 | minimal | Non | livré | [cap-cra-anomaly-factual.md](capabilities/cap-cra-anomaly-factual.md) |
| `cra.comment_summary` | 02 CRA | 2 | P2 | limité | Non | livré | [cap-cra-comment-summary.md](capabilities/cap-cra-comment-summary.md) |
| `budget.estimate` | 04 Budget | 2 | P1 | minimal | Non | livré | [cap-budget-estimate-effort.md](capabilities/cap-budget-estimate-effort.md) |
| `budget.demand_suggest` | 04 Budget | 2 | P1 | minimal | Non | livré | [cap-budget-demand-autocomplete.md](capabilities/cap-budget-demand-autocomplete.md) |
| `budget.overrun_forecast` | 04 Budget | 2 | P2 | minimal | Non | livré | [cap-budget-overrun-forecast.md](capabilities/cap-budget-overrun-forecast.md) |
| `dashboard.briefing` | transversal | 2 | P1 | minimal/limité | Non | livré | [cap-dashboard-briefing.md](capabilities/cap-dashboard-briefing.md) |
| `conges.date_suggest` | 03 Congés | 3 | P2 | limité | Non | livré | [cap-conges-date-suggest.md](capabilities/cap-conges-date-suggest.md) |
| `conges.manager_assist` | 03 Congés | 3 | P2 | haut potentiel | §4(b) | livré | [cap-conges-manager-assist.md](capabilities/cap-conges-manager-assist.md) |
| `workflow.explain` | 01 Workflow | 3 | P2 | minimal | Non | livré | [cap-workflow-explain.md](capabilities/cap-workflow-explain.md) |
| `publicsite.chatbot` | 15 Public | 3 | P2 | limité Art. 50 | Non | livré | [cap-publicsite-chatbot.md](capabilities/cap-publicsite-chatbot.md) |
| `publicsite.lead_scoring` | 15 Public | 3 | P3 | minimal | Non | livré | [cap-publicsite-lead-scoring.md](capabilities/cap-publicsite-lead-scoring.md) |
| `notifications.digest` | 11 Notif | 3 | P3 | limité | Non | livré | [cap-notifications-digest.md](capabilities/cap-notifications-digest.md) |
| `mobile.voice_cra` | 16 Mobile | 1bis | P2 | limité | Non | livré | [cap-mobile-voice-cra.md](capabilities/cap-mobile-voice-cra.md) |

## Capabilities exclues (IA Act)

| Cas d'usage | Motif |
| --- | --- |
| Anomalies CRA vs moyenne équipe | Monitoring performance — Annexe III §4(b) |
| Scoring / classement collaborateurs | Art. 5 / Annexe III |
| Validation auto congés / CRA / gate TMA | Art. 14 — supervision humaine obligatoire |
| Reconnaissance émotions au travail | Art. 5(1)(f) interdit |

## Évaluations Art. 6(3)

### `tma.suggest_assignee`

- **Annexe III** : allocation tâches basée sur historique individuel — potentiel §4(b).
- **Évaluation** : **non haut-risque** si :
  - Suggestion uniquement (pas d'assignation auto)
  - Basée sur charge et compétences déclarées, pas sur traits personnels
  - Manager confirme manuellement via select existant
- **Décision** : limité avec garde-fous — activation Vague 2 après revue.

### `conges.manager_assist`

- **Annexe III** : aide décision affectant relation de travail — §4(b).
- **Évaluation** : **non haut-risque** si :
  - Contexte factuel uniquement (soldes, absences équipe, pas de recommandation approve/reject)
  - Manager décide toujours via boutons existants
  - Info travailleurs requise avant activation tenant
- **Décision** : livrable Vague 3 avec explicabilité Art. 86.

### `mobile.voice_cra` (roadmap)

- Consentement explicite, pas de profilage — limité.

## Seed capabilities DB

Les codes ci-dessus sont insérés via migration `ai/0001_init.up.sql` dans `ai.ai_capabilities`.

## Revue périodique

- **Fréquence** : trimestrielle ou à chaque nouvelle capability
- **Responsable** : équipe produit + revue juridique si haut risque
- **Artefact** : mise à jour de ce fichier + fiches `cap-*.md`

## Definition of Done

- [ ] Matrice à jour à chaque nouvelle capability
- [ ] Évaluations Art. 6(3) pour capabilities ambiguës
- [ ] Sync avec table `ai.ai_capabilities`
