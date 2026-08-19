# Brique 18 — Module PROJET (planification agile)

> Planification agile par **Application** : profil méthodologique (`psa` | `agile_scrum` | `agile_kanban`), Epics, Sprints, Backlog, métriques.

## Périmètre

- **Inclus** : schéma `project`, API `/applications/{id}/epics|sprints|backlog|velocity`, UI `/projets/*`, terminologie dynamique via `useMethodologyTerms`.
- **Pivot inchangé** : `tma.demands` reste l'entité opérationnelle (CRA, budget, workflow) ; champs agile `epic_id`, `sprint_id`, `story_points`, `backlog_rank`.

## Dépendances

- 00 Organisation (`Application.methodologyProfile`)
- 05 TMA (`Demand` enrichie)
- 01 Workflow (preset `tma.incident` réutilisé)

## DoD

- [ ] Migration `project` + `org.0031` + `tma.0004`
- [ ] RBAC module `project` + entitlement billing
- [ ] OpenAPI + tests domain/app
- [ ] UI backlog/sprints/metrics + i18n fr/en
