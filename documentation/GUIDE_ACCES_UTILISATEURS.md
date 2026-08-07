# Guide d'accès utilisateurs (RBAC)

> **In-app** : section **Aide → Accès par profil** (`/aide`, `/aide/acces`).  
> **Dernière mise à jour doc** : 2026-08-07  
> **Sources de vérité code** :
> - `internal/modules/org/app/service.go` → `DefaultPermissions()`
> - `frontend/utils/rbac.ts` → `PROFILE_PERMISSIONS` (miroir)
> - Navigation : `frontend/layouts/default.vue`

Ce document décrit les **droits MVP réellement branchés**. La matrice cible complète reste dans
[`SPECIFICATION_FONCTIONNELLE.md`](SPECIFICATION_FONCTIONNELLE.md) §3.

## Obligation de mise à jour

Toute évolution des permissions, menus ou middlewares d'accès doit mettre à jour **dans la même PR** :

1. Ce fichier (`documentation/GUIDE_ACCES_UTILISATEURS.md`)
2. Les textes i18n `help.*` (`frontend/locales/fr.json`, `en.json`)
3. Le miroir front `frontend/utils/rbac.ts` si la matrice change
4. `DefaultPermissions()` côté API si la matrice change
5. Tests : Vitest `formatRbacCell` + Playwright `frontend/e2e/smoke/aide.spec.ts` si parcours Aide impacté

Règle agent : `.cursor/rules/guide-acces-doc.mdc`.

## Légende

| Code | Signification |
| --- | --- |
| **L** | Lecture |
| **E** | Écriture |
| **V** | Validation |
| **—** | Pas d'accès |

Un utilisateur peut cumuler plusieurs profils : les droits sont l'**union** (le plus large gagne).

## Matrice MVP (profil × module)

| Profil | Org | CRA | TMA | Congés | Budget | Reporting | Support | Maint. | Workflow | Billing | Notif. | Intégr. | Factur. | Admin | SSII | ETT |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Administrateur | L/E/V | L/E/V | L/E/V | L/E/V | L/E/V | L | L/E/V | L/E/V | L/E/V | L/E | L/E | L/E/V | L/E/V | L/E/V | L/E/V | L/E/V |
| Collaborateur | — | L/E | L/E | L/E | L | — | — | — | — | — | — | — | — | — | — | — |
| Chef d'équipe | L | L/E/V | L/E/V | L | L/E | L | — | — | — | — | — | — | L/E/V | — | — | — |
| Responsable de service | L | L/E/V | L/E/V | L/E/V | L/E/V | L | — | — | — | — | — | — | L/E/V | — | — | — |

## Fiches profils (comportement UI)

### Administrateur

- Menus admin : organisation (**identité / modules / structure**), **applications**, utilisateurs, SSO, workflows, paramètres, abonnement, notifications.
- Organisation → Modules : toggle **facturation client** (`invoicing_enabled`) — indépendant de l'abonnement SaaS Stripe ; le menu Facturation n'apparaît que si activé.
- Page `/admin/applications` : CRUD applications (libellé, propriétaire, mode facturation, UO, chef utilisateur), équipes liées, vue users/budgets ; désactivation soft.
- Validation complète CRA / TMA / congés / budget.
- Facturation métier : L/E/V si le module org est activé.
- CRA définitifs → brouillons facture :
  - **Facturation → Depuis un CRA** (wizard preview prérempli, droits Facturation écriture **et** CRA validation) ;
  - validation définitive CRA ou Prestations → **Créer factures** (mêmes droits).
- Proforma : émission email client avec lien de validation publique ; à validation, conversion en facture préparée + envoi auto de la facture (PDP reste manuel).
- Heures facturées bornées à la mission (ou app `temps_passe` dominante) — pas d’agrégat multi-clients.
- Self-edit (`/admin/users`) : peut cumuler d'autres profils / équipes, **ne peut pas** retirer son propre profil Administrateur ni désactiver son compte ; le dernier Administrateur actif du tenant est protégé.
- Comptes seed (visibles dans l'Aide **uniquement** pour les Administrateurs) : `ADM_admin` / `Admin123!`

### Collaborateur

- CRA, TMA, congés (saisie) ; budget en lecture.
- Pas de menus d'administration.
- Compte seed : `COL_collab` / `Collab123!`

### Chef d'équipe

- Validation CRA et TMA ; congés en lecture ; budget L/E ; reporting L.
- Facturation métier : L/E/V si le module org est activé (wizard **Depuis un CRA**, Prestations, proforma client, transmission PDP).
- Compte seed : `CHE_chefdev` / `Chef123!`

### Responsable de service

- Validation CRA, TMA, congés et budget ; reporting / org en lecture.
- Facturation métier : L/E/V si le module org est activé (wizard **Depuis un CRA**, Prestations, proforma client, transmission PDP).
- Pas d'accès aux paramètres admin (réservé Administrateur).
- Compte seed : `MGR_manager` / `Manager123!`

## Profils prévus (non branchés dans DefaultPermissions)

| Profil | Intention (SFD §3) |
| --- | --- |
| Utilisateur | Demandes / incidents canal utilisateur |
| Commercial | Clients, missions, facturation prévisionnelle |
| Support | Tickets helpdesk |
| Chef utilisateur | Gate validation amont TMA |
| Client externe | Incidents / suivi côté client |
| Sous-traitant | Consommation / CRA prestataire |

## Navigation Aide

| Route | Contenu |
| --- | --- |
| `/aide` | Hub d'aide (profils courants + topics) |
| `/aide/acces` | Matrice L/E/V + fiches profils |

Entrée nav : **Aide** (icône `help`), zone réglages, visible pour tout utilisateur authentifié.
