# cap-tma-analysis-draft — Brouillon dossier analyse TMA

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `tma.analysis_draft` |
| Module | 05 TMA |
| Vague | 1 — P0 |
| Statut | livré |

## Finalité métier

Générer un brouillon des 4 champs analyse (fonctionnel, technique, risques, tests) à partir du sujet demande + contexte application. Persona : développeur TMA.

## Classification IA Act

- **Risque** : minimal / limité (contenu généré Art. 50)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous RG-TMA-01

- Ne modifie pas le statut demande ni la gate chef utilisateur
- Brouillon **non enregistré** tant que l'utilisateur n'a pas cliqué « Enregistrer »
- Validation humaine obligatoire avant résolution workflow

## Entrées / sorties

**Entrée** : `demandId`, sujet, `applicationId`  
**Sortie** : `{ functional, technical, risks, testScenario, requestId }`

## Ancrage code

- [`frontend/components/tma/AnalysisEditor.vue`](../../../frontend/components/tma/AnalysisEditor.vue)
- [`frontend/pages/tma/[id].vue`](../../../frontend/pages/tma/[id].vue)
- `internal/modules/ai/app/service.go` — `SuggestAnalysisDraft`

## API

- BFF : `POST /api/ai/tma/analysis-draft`
- Go : `POST /api/v1/ai/tma/analysis-draft`

## UX

- Bouton « Générer brouillon » + `AppAiBadge` + disclaimer
- Pattern accept : remplit champs locaux, reject : ignore

## DoD

- [x] Route BFF + handler Go
- [x] Journalisation `ai_request_log`
- [x] Badge UI + i18n
