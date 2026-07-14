# cap-tma-executive-summary — Synthèse exécutive TMA

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `tma.executive_summary` |
| Module | 05 TMA |
| Vague | 2 — P2 |
| Statut | livré |

## Finalité métier

Générer une **synthèse textuelle** du portefeuille TMA (ouverts, résolus, délais, charge) pour reporting manager. Persona : responsable TMA / direction technique.

## Classification IA Act

- **Risque** : limité (contenu généré Art. 50)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- Agrégats **factuels** injectés dans le prompt (KPIs, pas de données nominatives)
- Synthèse en lecture seule — pas d'action workflow
- Disclaimer + badge « généré par IA »

## Entrées / sorties

**Entrée** : `period`, filtres optionnels (`applicationId`, `status`)  
**Sortie** : `{ summary, highlights[], requestId }`

## Ancrage code

- [`frontend/pages/tma/index.vue`](../../../frontend/pages/tma/index.vue)
- [`frontend/composables/useKpiMetrics.ts`](../../../frontend/composables/useKpiMetrics.ts)
- `internal/modules/ai/app/service.go` — `GenerateTmaExecutiveSummary` (à créer)

## API

- BFF : `GET /api/ai/tma/executive-summary`
- Go : `GET /api/v1/ai/tma/executive-summary`

## DoD

- [ ] Route BFF + handler Go
- [ ] Journalisation `ai_request_log`
- [ ] Carte synthèse + `AppAiBadge` sur liste TMA
- [ ] i18n fr/en
