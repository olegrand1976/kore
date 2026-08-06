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
  - en GitHub Actions via **évaluation IA (Gemini)** sur les commits depuis le dernier tag
    (`GEMINI_API_KEY` / `GEMINI_MODEL`, comme le reste du projet — pas d’OpenAI)
  - **fallback déterministe Conventional Commits** si la clé est absente ou l’appel échoue.
- **Sorties** : création du tag `vX.Y.Z` (et éventuellement GitHub Release + notes).

## Checklist avant PR UI

1. Mobile ≤768px : nav drawer + bottom bar app, pas de sidebar seule
2. CTAs pleine largeur sur mobile
3. i18n : zéro string FR hardcodé dans templates
4. Tokens charte, pas de couleurs ad hoc
5. `npm run build` + `go build ./...`
6. **Flux de tests** : Vitest et/ou Playwright smoke étendus pour la feature (même PR) — règle `feature-tests-sync`
7. Si parcours UI critique : spec Playwright (`frontend/e2e/smoke/`) ou extension d'une existante
8. Logique composable/BFF : test Vitest (`frontend/tests/`)
9. Si RBAC / menus changent : maj Aide (`/aide`, i18n `help.*`) + `documentation/GUIDE_ACCES_UTILISATEURS.md`

## Checklist avant PR backend (migrations)

1. Migration `.up.sql` + test d'intégration si pertinent
2. **`documentation/SCHEMA_DB.md` à jour** (même PR que la migration)
3. Si `DefaultPermissions` / profils changent : maj `frontend/utils/rbac.ts` + guide d'accès (Aide + doc)
4. **Flux de tests** : domain/app (+ intégration / smoke API si impact) dans la même PR — règle `feature-tests-sync`
5. `go test ./...` + `make migrate`
6. Nouvelle règle métier → test domain ; nouveau use case → test app (fakes ports)
7. Alter SQL → `*_integration_test.go` (round-trip ou contrainte)

## Tests (filet anti-régression)

Toute nouvelle feature **doit** étendre ce filet (pas de livraison « tests plus tard »).

```bash
make test              # unitaires Go
make test-frontend     # Vitest
make test-integration  # Postgres (Docker)
make test-e2e          # Playwright smoke (stack Docker test)
make test-all          # pyramide complète
```

CI : jobs `backend`, `frontend`, `integration`, `e2e`, `smoke` — voir `technical/foundation/06-testing-strategy.md`.

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
| Aide / accès | `frontend/pages/aide/`, `documentation/GUIDE_ACCES_UTILISATEURS.md`, `frontend/utils/rbac.ts` |
| Wiki GitHub | sync auto au deploy (`scripts/sync-github-wiki.sh`) |
