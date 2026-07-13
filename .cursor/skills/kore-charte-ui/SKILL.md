---
name: kore-charte-ui
description: >-
  Charte graphique Kore — tokens CSS, composants Public/App, logo hero PNG,
  thème dark/light. Use when creating or modifying UI pages, components,
  layouts, or reviewing visual consistency.
---

# Kore — charte UI

## Tokens

`frontend/assets/css/tokens.css` — `--kore-*` uniquement, jamais de hex ad hoc.

### Largeur layout (desktop)

| Token | Valeur | Usage |
|-------|--------|-------|
| `--kore-app-main-max` | `none` | App authentifiée : `<main>` fluide sur toute la colonne |
| `--kore-public-container-max` | `1440px` | Site public : layout + footer |
| `--kore-prose-max` | `720px` | Titres/intro marketing (`.page-hero__inner`) |
| `--kore-form-max` | `420px` | Champs formulaire (pas le layout page) |
| `--kore-form-wide-max` | `480px` | Formulaires plus larges (TMA, modales) |
| `--kore-container-max` | `1200px` | Legacy — `definePageMeta({ narrow: true })` app |

## Logo

Source `logo/kore logo.png` → `<KoreLogo variant="hero" />`.

## Composants

Public* pour surface publique, App* pour app authentifiée.

## Nouvel écran — règles largeur

1. **App** : pas de `max-width` sur le wrapper page — le layout est fluide par défaut
2. **Public** : contenu dans `.page-shell` ; prose hero limitée à `--kore-prose-max`
3. **Formulaires** : limiter les champs via `--kore-form-max`, pas le `<main>`
4. **Tableaux/listes** : `AppTable` / grilles à 100% du parent
5. **`definePageMeta({ narrow: true })`** : uniquement pages centrées type wizard
6. **Mobile ≤768px** : grilles 1 col, CTAs full-width, bottom nav — inchangé

## Checklist

- [ ] Tokens + i18n FR/EN
- [ ] Logo alt
- [ ] Responsive mobile
- [ ] Empty states + feedback erreurs
- [ ] Desktop : listes/tableaux utilisent la largeur disponible
