---
name: kore-responsive-check
description: >-
  Vérifie et corrige le responsive mobile des écrans Kore (320–768px).
  Use after creating or modifying Vue pages, layouts, or CSS — or when
  the user mentions mobile, responsive, or viewport.
---

# Kore — contrôle responsive

## Procédure

1. Layout : `public.vue` (drawer menu) ou `default.vue` (bottom nav + drawer)
2. `AppPageHeader` stack ≤640px
3. Grilles 1 colonne, CTAs full-width mobile
4. `padding-bottom` main pour bottom nav + safe-area
5. `npm run build`

## Breakpoints

- mobile : max 768px
- mobile-sm : max 640px
- desktop : min 900px

## Rapport

✅ OK / ⚠️ corrigé / ❌ bloquant par écran.
