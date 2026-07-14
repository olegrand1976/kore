# cap-publicsite-lead-scoring — Scoring leads publicsite

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `publicsite.lead_scoring` |
| Module | 15 Public |
| Vague | 3 — P3 |
| Statut | livré |

## Finalité métier

Attribuer un **score d'intention** aux demandes de démo / contact (taille ESN, modules cités, complétude formulaire) pour priorisation commerciale interne. Persona : équipe commerciale Kore.

## Classification IA Act

- **Risque** : minimal (prospects B2B, pas de scoring collaborateurs)
- **Annexe III** : Non — exclut scoring social Art. 5
- **Art. 6(3)** : N/A

## Garde-fous

- Score **interne** — non visible du prospect
- Basé sur champs formulaire déclarés, pas de profilage comportemental invasif
- Consentement RGPD formulaire contact obligatoire

## Entrées / sorties

**Entrée** : payload formulaire (`reserver`, contact)  
**Sortie** : `{ score, tier, factors[], requestId }`

## Ancrage code

- [`frontend/pages/reserver/index.vue`](../../../frontend/pages/reserver/index.vue)
- [`frontend/pages/tarifs/index.vue`](../../../frontend/pages/tarifs/index.vue)
- `internal/modules/ai/app/service.go` — `ScorePublicLead` (à créer)

## API

- BFF : `POST /api/ai/public/lead-score`
- Go : `POST /api/v1/ai/public/lead-score`

## DoD

- [ ] Route BFF + handler Go (auth interne uniquement)
- [ ] Journalisation `ai_request_log`
- [ ] Affichage score côté admin commercial (roadmap)
- [ ] Revue conformité Art. 5 avant activation
