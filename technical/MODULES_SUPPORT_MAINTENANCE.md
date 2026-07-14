# Support (M06) et Maintenance (M07) — plan de continuité

> Squelettes livrés dans `internal/modules/support/` et `internal/modules/maintenance/`.
> Complète [ROADMAP.md](ROADMAP.md) post-MVP.

## État actuel

| Module | Package | Backend | Frontend | Priorité |
| --- | --- | --- | --- | --- |
| **06 Support** | `support/` | Routes HTTP + schéma `support` | Pages `support/` + paramètres demandes | Après stabilisation TMA |
| **07 Maintenance** | `maintenance/` | Routes HTTP + schéma `maintenance` | Pages `maintenance/` + paramètres demandes | Après M06 |

Les canaux TMA / Service utilisateur / Exploitation & travaux sont activables par tenant via `org.tenant_request_settings` (admin → Paramètres → Demandes).

## Dépendances

- **01 Workflow** — cycle ticket/travaux
- **05 TMA** — modèle Demande partagé (artefacts allégés pour M07)
- **02 CRA** — alimentation à la résolution/terminaison
- **11 Notifications** — réponses historisées (RG-SUP-01)

## Séquençage recommandé

1. **M06 Support** (6–8 sem.) — tickets web + ingestion mail, réponses, résolution, CRA feeder
2. **M07 Maintenance** (4–6 sem.) — cycle allégé Créé → En cours → Terminé, réutilise workflow TMA

## Prochaines étapes

- [ ] Intégration `InboundMailGateway` (M06)
- [x] UI Nuxt `support/` et `maintenance/`
- [x] Paramètres admin activation canaux (`/admin/parametres/demandes`)
- [ ] Tests app + intégration workflow
- [ ] Entitlements module 14 pour activation tenant
