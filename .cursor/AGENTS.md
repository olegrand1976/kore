# Kore — Agents Cursor

Monolithe PSA/ESN : API Go (`internal/`) + frontend Nuxt 3 (`frontend/`).

## Sub-agents recommandés

| Agent | Rôle | Quand l'utiliser |
|-------|------|------------------|
| **kore-frontend** | Pages Vue, composants charte, BFF Nitro, i18n, responsive | Toute tâche `frontend/**` |
| **kore-backend** | Modules Go, migrations, handlers HTTP, tests | Toute tâche `internal/**`, `cmd/**` |
| **kore-fullstack** | Feature bout-en-bout (API + BFF + UI) | CRA, org, billing, auth |
| **kore-review** | Code review sécurité/UX/charte | Avant merge, après grosse feature |
| **kore-mobile-qa** | Vérification responsive 320–768px | Après tout écran UI |

## Stack & ports locaux

```bash
make up          # stack Docker
# Frontend : http://localhost:3001
# API      : http://localhost:8081
# Admin    : ADM_admin / Admin123!
```

## Conventions transverses

- **Langue user-facing** : FR par défaut, i18n `@nuxtjs/i18n` (`frontend/locales/`)
- **Charte** : tokens `--kore-*` dans `frontend/assets/css/tokens.css` — doc `documentation/CHARTE_GRAPHIQUE.md`
- **Logo source** : `logo/kore logo.png` → `frontend/public/brand/kore-logo-hero.png`
- **Auth app** : cookie httpOnly `kore_access_token`, session `/api/auth/session`, middleware `auth.global.ts`
- **RBAC nav** : admin = `profile === 'Administrateur'`, middleware `admin.ts`

## Release notes & versioning (conventions produit)

### Modale “Quoi de neuf” à la connexion

- **But** : afficher une modale après connexion avec les changements du projet, **commits groupés par mois** (select).
- **Source** : GitHub API **côté serveur** (BFF Nitro dans `frontend/server/api/**`) — ne jamais exposer de token GitHub au client.
- **Affichage** :
  - auto à la connexion si `last_seen_version != current_version` et si l’utilisateur n’a pas désactivé l’auto-affichage
  - bouton dans la topbar (layout app) pour ouvrir la modale manuellement.
- **Persistance** : préférences **par utilisateur** côté backend (DB) :
  - `release_notes_auto_show` (bool)
  - `last_seen_version` (texte/SemVer)

### Versioning automatique (CI)

- **Portée** : tags git SemVer **uniquement** (pas de bump `package.json`/`go.mod`).
- **Décision major/minor/patch** :
  - en GitHub Actions via **évaluation IA (OpenAI)** sur les commits depuis le dernier tag
  - prévoir un **fallback déterministe** si l’IA échoue (ex: Conventional Commits).
- **Sorties** : création du tag `vX.Y.Z` (et éventuellement GitHub Release + notes).

## Checklist avant PR UI

1. Mobile ≤768px : nav drawer + bottom bar app, pas de sidebar seule
2. CTAs pleine largeur sur mobile
3. i18n : zéro string FR hardcodé dans templates
4. Tokens charte, pas de couleurs ad hoc
5. `npm run build` + `go build ./...`

## Checklist avant PR backend (migrations)

1. Migration `.up.sql` + test d'intégration si pertinent
2. **`documentation/SCHEMA_DB.md` à jour** (même PR que la migration)
3. `go test ./...` + `make migrate`

## Skills projet (`.cursor/skills/`)

- `kore-dev-workflow` — commandes, structure modules, BFF
- `kore-charte-ui` — composants Public*/App*, tokens, logo
- `kore-responsive-check` — checklist mobile obligatoire

## Fichiers clés

| Domaine | Fichiers |
|---------|----------|
| Layouts | `frontend/layouts/public.vue`, `default.vue` |
| Mobile | `MobileDrawer.vue`, `AppBottomNav.vue` |
| Thème | `composables/useTheme.ts`, `tokens.css` |
| CRA | `internal/modules/cra/`, `frontend/pages/cra/` |
| Org/branding | `internal/modules/org/`, `frontend/pages/admin/organisation/` |
| Schéma DB | `documentation/SCHEMA_DB.md`, `internal/modules/*/migrations/` |
| Wiki GitHub | sync auto au deploy (`scripts/sync-github-wiki.sh`) |
