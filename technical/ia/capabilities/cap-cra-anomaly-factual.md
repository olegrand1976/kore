# cap-cra-anomaly-factual — Anomalies CRA factuelles

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `cra.anomalies` |
| Module | 02 CRA |
| Vague | 2 — P1 |
| Statut | livré |

## Finalité métier

Signaler des **écarts factuels** sur une feuille CRA (capacité jour, trous, doublons source, incohérence commercial) pour alerte manager. Persona : consultant / manager CRA.

## Classification IA Act

- **Risque** : minimal — **pas de LLM** (règles déterministes Go)
- **Annexe III** : Non — exclut monitoring performance vs moyenne équipe (voir registre)
- **Art. 6(3)** : N/A — capability enregistrée pour traçabilité, moteur = règles métier

## Garde-fous RG-CRA-01 / RG-CRA-03

- Règles **déterministes** uniquement : seuils capacité, continuité, doublons `source_id`
- Pas de comparaison inter-collaborateurs ni scoring
- Alerte informative — ne bloque pas la saisie (sauf règles métier existantes RG-CRA-03)

## Entrées / sorties

**Entrée** : `timesheetId`  
**Sortie** : `{ anomalies: [{ code, severity, day?, message }], requestId }`

## Ancrage code

- [`frontend/pages/cra/[id].vue`](../../../frontend/pages/cra/[id].vue)
- [`frontend/components/cra/TimesheetGrid.vue`](../../../frontend/components/cra/TimesheetGrid.vue)
- `internal/modules/cra/domain/domain.go` — règles capacité
- `internal/modules/ai/app/service.go` — `DetectCraAnomalies` (sans appel LLM)

## API

- BFF : `GET /api/ai/cra/anomalies`
- Go : `GET /api/v1/ai/cra/anomalies`

## DoD

- [x] Routes BFF + handlers Go
- [x] Moteur règles déterministes (zéro token LLM)
- [x] Bandeau alertes sur détail CRA
- [x] Journalisation `ai_request_log` (provider=rules)
