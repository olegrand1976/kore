# Charte graphique Kore

Référence visuelle pour le produit Kore (web, PDF, emails).

## Couleurs de marque (invariantes)

| Token | Hex | Usage |
|---|---|---|
| `--kore-brand-navy` | `#1e3a5f` | Titres, wordmark |
| `--kore-brand-blue` | `#2b6cb0` | Liens marque, emblème |
| `--kore-brand-gold` | `#c9a227` | CTA, accents premium |

## Thèmes

- **Dark** (défaut public) : fond `#1a1f2e`, texte `#e8eaed`
- **Light** : fond `#f8f9fb`, texte `#1a1f2e`
- Bascule : `ThemeToggle` + `localStorage` clé `kore-theme`

Tokens complets : [`frontend/assets/css/tokens.css`](../frontend/assets/css/tokens.css)

## Logo

| Variante | Fichier | Usage |
|---|---|---|
| Emblème couleur | `frontend/public/brand/kore-emblem.svg` | Favicon, sidebar |
| Horizontal | `kore-logo-horizontal.svg` | Header public |
| Full | `kore-logo-full.svg` | Hero light mode |
| Hero 3D | `kore-logo-hero.png` | Hero dark mode |
| Mono light/dark | `kore-emblem-mono-*.svg` | Sidebar selon thème |

Composant : `<KoreLogo variant="horizontal" tone="auto" />`

### Règles

- Hauteur header : `sm` (120px horizontal) max
- Toujours `alt="Kore"` ou raison sociale tenant
- Ne pas déformer le ratio

## Tenant branding

- Upload : `/admin/organisation` → API `PUT /api/v1/societes/{id}/branding`
- Fallback : emblème Kore si pas de logo
- PDF : header société + footer « Généré par Kore » (masquable entitlement marque blanche)

## PDF (CRA)

Template HTML charté : `internal/modules/cra/adapters/pdf/templates/cra.html`

Couleurs alignées charte web. Export MIME `text/html` (impression navigateur / conversion PDF ultérieure).

## Emails

Template base : `internal/modules/notifications/adapters/email/templates/base.html`

Signature : Cordialement + société + URL tenant (spec §12.2).

## Stripe Checkout

Branding session : primary `#c9a227`, background `#1a1f2e`, logo horizontal Kore.

## i18n

FR (défaut) + EN via `@nuxtjs/i18n`. Wordmark **KORE** invariant.

## Composants UI

| Public | App |
|---|---|
| `PublicButton`, `PublicCard`, `PublicInput` | `AppButton`, `AppCard`, `AppInput`, `AppTable`, `AppBadge` |
| `PublicSection`, `PublicFooter` | `AppPageHeader`, `TenantLogo` |

## Checklist revue PR

- [ ] Aucune couleur hardcodée hors tokens
- [ ] Contraste WCAG AA dark + light
- [ ] Logo avec alt
- [ ] ThemeToggle fonctionnel sans flash
- [ ] Pages publiques + app cohérentes
