# cap-notifications-digest — Digest notifications IA

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `notifications.digest` |
| Module | 11 Notifications |
| Vague | 3 — P3 |
| Statut | planifié |

## Finalité métier

Synthétiser les **notifications non lues** en un résumé quotidien ou hebdomadaire actionnable (TMA, congés, CRA, workflow). Persona : utilisateur authentifié.

## Classification IA Act

- **Risque** : limité (contenu généré Art. 50)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous

- Agrégation **par utilisateur** — isolation tenant stricte
- Digest = lecture — liens vers actions manuelles, pas d'exécution auto
- Opt-in utilisateur distinct de l'activation tenant IA

## Entrées / sorties

**Entrée** : `userId`, `period` (`daily` | `weekly`)  
**Sortie** : `{ digest, itemCount, links[], requestId }`

## Ancrage code

- [`frontend/pages/admin/notifications/index.vue`](../../../frontend/pages/admin/notifications/index.vue)
- [`frontend/layouts/default.vue`](../../../frontend/layouts/default.vue)
- `internal/modules/ai/app/service.go` — `GenerateNotificationDigest` (à créer)

## API

- BFF : `GET /api/ai/notifications/digest`
- Go : `GET /api/v1/ai/notifications/digest`

## DoD

- [ ] Route BFF + handler Go
- [ ] Journalisation `ai_request_log`
- [ ] Préférence utilisateur digest on/off
- [ ] i18n fr/en + `AppAiBadge`
