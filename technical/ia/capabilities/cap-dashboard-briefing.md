# cap-dashboard-briefing — Briefing tableau de bord

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `dashboard.briefing` |
| Module | transversal |
| Vague | 2 — P1 |
| Statut | livré |

## Finalité métier

Générer un **briefing personnalisé** (2–4 phrases) résumant l'activité du jour : CRA en cours, validations congés, TMA ouverts, alertes budget. Persona : tout utilisateur authentifié.

## Classification IA Act

- **Risque** : minimal / limité (contenu généré Art. 50)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- KPIs injectés en **données structurées** — pas de PII tierces dans le prompt
- Cache TTL 5 min par utilisateur — pas de régénération à chaque navigation
- Lecture seule — aucune action métier déclenchée

## Entrées / sorties

**Entrée** : contexte session (`userId`, modules actifs)  
**Sortie** : `{ briefing, generatedAt, requestId }`

## Ancrage code

- [`frontend/pages/dashboard/index.vue`](../../../frontend/pages/dashboard/index.vue)
- [`frontend/composables/useDashboardStats.ts`](../../../frontend/composables/useDashboardStats.ts)
- [`frontend/composables/useKpiMetrics.ts`](../../../frontend/composables/useKpiMetrics.ts)
- `internal/modules/ai/app/service.go` — `GenerateDashboardBriefing`

## API

- BFF : `GET /api/ai/dashboard/briefing`
- Go : `GET /api/v1/ai/dashboard/briefing`

## DoD

- [x] Routes BFF + handlers Go
- [x] Carte briefing + `AppAiBadge` + `aria-live="polite"`
- [x] Cache Redis / stub TTL 5 min
- [x] Journalisation `ai_request_log`
