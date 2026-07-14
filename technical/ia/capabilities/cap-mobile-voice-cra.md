# cap-mobile-voice-cra — Saisie vocale CRA mobile

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `mobile.voice_cra` |
| Module | 16 Mobile |
| Vague | 1bis — P2 |
| Statut | livré |

## Finalité métier

Permettre la **dictée vocale** d'une ligne CRA (jour, durée, commentaire) sur mobile, convertie en proposition de saisie modifiable. Persona : consultant en déplacement.

## Classification IA Act

- **Risque** : limité (transcription + structuration)
- **Annexe III** : Non
- **Art. 6(3)** : N/A — consentement explicite, pas de profilage

## Garde-fous RG-CRA-01

- Transcription → **brouillon local** — validation humaine avant enregistrement
- Lignes `origin=manual` jamais écrasées
- Consentement micro explicite (navigateur / PWA)
- Pas de stockage audio brut côté serveur (transcription uniquement)

## Entrées / sorties

**Entrée** : `timesheetId`, `week`, `audio` ou `transcript`  
**Sortie** : `{ lines: [{ day, duration, comment }], requestId }`

## Ancrage code

- [`frontend/pages/cra/[id].vue`](../../../frontend/pages/cra/[id].vue)
- [`frontend/components/cra/WeekEditor.vue`](../../../frontend/components/cra/WeekEditor.vue)
- [`frontend/components/ui/AppBottomNav.vue`](../../../frontend/components/ui/AppBottomNav.vue)
- `internal/modules/ai/app/service.go` — `TranscribeCraVoice` (à créer)

## API

- BFF : `POST /api/ai/mobile/voice-cra`
- Go : `POST /api/v1/ai/mobile/voice-cra`

## DoD

- [ ] Route BFF + handler Go
- [ ] UI micro mobile ≤768px + consentement
- [ ] Pattern accept/reject avant `saveWeek`
- [ ] Journalisation `ai_request_log`
- [ ] Test responsive 320px / 768px
