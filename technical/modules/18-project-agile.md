# Brique 18 — Module PROJET (planification agile)

> Planification agile par **Application** : profil méthodologique (`psa` | `agile_scrum` | `agile_kanban`), Epics, Sprints, Backlog, métriques.

## Périmètre

- **Inclus** : schéma `project`, API `/applications/{id}/epics|sprints|backlog|velocity`, UI `/projets/*`, terminologie dynamique via `useMethodologyTerms`.
- **Pivot inchangé** : `tma.demands` reste l'entité opérationnelle (CRA, budget, workflow) ; champs agile `epic_id`, `sprint_id`, `story_points`, `backlog_rank`, `resolved_at`.

## Dépendances

- 00 Organisation (`Application.methodologyProfile`)
- 05 TMA (`Demand` enrichie)
- 01 Workflow (preset `tma.incident` réutilisé)

## DoD

- [x] Migration `project` + `org.0031` + `tma.0004` + `tma.0005` (`resolved_at`)
- [x] RBAC module `project` + entitlement billing
- [x] OpenAPI + tests domain/app/integration + smoke API + Playwright `/projets`
- [x] UI backlog (drag), sprints (plan), board Kanban + config WIP, métriques burndown, champs agile TMA, i18n fr/en
- [x] Seed démo (Scrum + Kanban) + aide in-app `/aide/projets`
