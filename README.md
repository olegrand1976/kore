# Kore

Suite PSA/ESN unifiée autour du **CRA pivot** (temps → projet → facturation → RH), en reprise fonctionnelle de l'offre historique **B-Hive** (Bee Software, 2008–2013).

## Stack technique (décision actée — greenfield)

| Couche | Technologie |
| --- | --- |
| API & logique métier | **Go** (chi, pgx, golang-migrate) — monolithe modulaire hexagonal |
| Base de données | **PostgreSQL** (schéma par module) |
| Cache & sessions | **Redis** |
| Frontend | **Nuxt 3** (Vue 3, SSR + BFF Nitro) |
| Paiements SaaS | **Stripe** (abonnements) |
| Conteneurisation | **Docker** / Docker Compose (dev), **GCP** (prod) |

> Le dépôt Kore ne contient **aucun code legacy** PHP/Flash/Flex. La modernisation technique est une **réécriture greenfield** ; les sources B-Hive servent uniquement de référence fonctionnelle.

## Documentation

| Document | Rôle |
| --- | --- |
| [`documentation/SPECIFICATION_FONCTIONNELLE.md`](documentation/SPECIFICATION_FONCTIONNELLE.md) | Spécification fonctionnelle (modules, processus, règles) |
| [`documentation/ANALYSE_COMMERCIALE.md`](documentation/ANALYSE_COMMERCIALE.md) | Analyse commerciale et go-to-market |
| [`technical/README.md`](technical/README.md) | Spécifications techniques, architecture, briques |
| [`documentation/CHARTE_GRAPHIQUE.md`](documentation/CHARTE_GRAPHIQUE.md) | Charte visuelle UI |

## Démarrage local

```bash
make up    # frontend :3001, API :8081, PostgreSQL, Redis
```

Identifiants seed : `ADM_admin` / `Admin123!`

## Structure du dépôt

```
cmd/kore-api/          # Point d'entrée API Go
internal/modules/      # Modules métier hexagonaux (org, cra, tma, …)
frontend/              # Nuxt 3 + BFF (frontend/server/api/)
db/migrations/         # Migrations transverses
technical/             # Spécifications techniques détaillées
documentation/         # Spécifications fonctionnelles et commerciales
```
