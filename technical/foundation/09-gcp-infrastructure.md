# 09 — Infrastructure GCP & déploiement

> Fondation transverse. Cible de déploiement **Google Cloud Platform**. Le développement et les tests restent en **Docker Compose local** (cf. [07-docker-devops.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/07-docker-devops.md)).

## 1. Vue d'ensemble

- **Compute** : **Cloud Run** (conteneurs managés, scale-to-zero, autoscaling) — un service pour l'API Go, un service pour le frontend Nuxt (SSR/BFF).
- **Base de données** : **Cloud SQL for PostgreSQL** (managé, un schéma par module — cf. [03-database.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/03-database.md)).
- **Cache** : **Memorystore for Redis**, **instance partagée** avec d'autres applications — isolation par préfixe de clés (cf. [10-cache-redis.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/10-cache-redis.md)).
- **Secrets** : **Secret Manager** (clés JWT, mots de passe DB, clés Stripe, config PDP).
- **Images** : **Artifact Registry**.
- **CI/CD** : **Cloud Build** (build + push + déploiement Cloud Run) déclenché sur merge.
- **Réseau** : **VPC + Serverless VPC Access Connector** pour joindre Cloud SQL et Memorystore en **IP privée**.

```mermaid
flowchart TB
  subgraph Internet
    User["Navigateur"]
  end
  subgraph GCP["GCP - projet Kore"]
    LB["HTTPS Load Balancer / domaine tenant"]
    subgraph run["Cloud Run"]
      FE["Service frontend (Nuxt SSR/BFF)"]
      API["Service API (Go chi)"]
    end
    Conn["Serverless VPC Access Connector"]
    subgraph vpc["VPC (IP privée)"]
      SQL[("Cloud SQL PostgreSQL")]
      Redis[("Memorystore Redis (partagé)")]
    end
    SM["Secret Manager"]
    AR["Artifact Registry"]
  end
  User --> LB --> FE --> API
  API --> Conn
  Conn --> SQL
  Conn --> Redis
  API --> SM
  FE --> SM
```

## 2. Cloud Run

| Paramètre | API Go | Frontend Nuxt |
| --- | --- | --- |
| Concurrence | 80 (I/O bound) | 80 |
| Min instances | 1 (évite le cold start métier) | 0/1 selon trafic |
| CPU | 1 vCPU (ajustable) | 1 vCPU |
| Port | `$PORT` (injecté par Cloud Run) | `$PORT` |
| Connexion DB | Cloud SQL via **connecteur intégré** (socket `/cloudsql/...`) ou IP privée via connector | via API uniquement |
| VPC egress | connector (privé) pour SQL/Redis | connector si nécessaire |

- Le service API lit sa configuration depuis les variables d'environnement, secrets montés via Secret Manager.
- **Stateless** : aucun état local (les sessions/révocations vont dans Redis — cf. [04-auth-rbac.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/04-auth-rbac.md)), ce qui rend l'autoscaling sûr.

## 3. Cloud SQL for PostgreSQL

- Instance régionale, **haute disponibilité** (failover) en production.
- Connexion depuis Cloud Run :
  - **Cloud SQL Auth Proxy** intégré (recommandé) : chiffrement + IAM, connexion via socket unix `/cloudsql/PROJECT:REGION:INSTANCE`.
  - ou **IP privée** via le VPC connector.
- Authentification : mot de passe applicatif (Secret Manager) ou **IAM database authentication**.
- Pool : `pgxpool` dimensionné en tenant compte de `max_connections` de l'instance **et** du nombre max d'instances Cloud Run (concurrence × instances ≤ connexions dispo — sinon activer un pooler).
- Sauvegardes automatiques + PITR ; migrations appliquées via job dédié (cf. §6).

## 4. Memorystore for Redis (partagé)

- Instance **partagée avec d'autres applications** GCP → voir la stratégie complète dans [10-cache-redis.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/10-cache-redis.md).
- Points d'infrastructure :
  - Accès en **IP privée** via le VPC connector.
  - **Interdiction des commandes destructrices/globales** (`FLUSHALL`, `FLUSHDB`, `KEYS *`) : elles impacteraient les autres applications.
  - Politique d'éviction gérée au niveau de l'instance (partagée) : ne pas présumer de persistance, tout doit être **reconstructible** (cache-aside).
  - TLS activé si l'instance l'exige.

## 5. Secret Manager & configuration

| Secret | Usage |
| --- | --- |
| `DATABASE_URL` / credentials Cloud SQL | Accès DB |
| `JWT_SIGNING_KEY` | Signature JWT |
| `REDIS_ADDR` / `REDIS_AUTH` | Accès Memorystore |
| `REDIS_KEY_PREFIX` | Préfixe d'isolation (ex. `kore`) |
| `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET` | Paiements ([11-payments-stripe.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/11-payments-stripe.md)) |
| `PDP_*` | Connecteur PDP (module 09) |

- Aucun secret dans l'image ou le dépôt. Montés comme variables d'environnement ou volumes par Cloud Run.

## 6. CI/CD (Cloud Build)

```mermaid
flowchart LR
  Git["Merge sur main"] --> CB["Cloud Build"]
  CB --> Test["Lint + tests (unit + integration)"]
  Test --> Img["Build images -> Artifact Registry"]
  Img --> Migrate["Job migrations Cloud SQL (golang-migrate)"]
  Migrate --> Deploy["Deploy Cloud Run (API + frontend)"]
  Deploy --> Smoke["Smoke test /health /ready"]
```

- Les **migrations** s'exécutent dans un **job dédié** (Cloud Run Job ou étape Cloud Build) **avant** le basculement de trafic, jamais au boot en production (cf. [03-database.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/03-database.md) et [07-docker-devops.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/07-docker-devops.md)).
- Déploiement progressif possible (revisions Cloud Run + répartition de trafic).

## 7. Multi-tenant sur GCP

- Isolation logique par `tenant_id` (cf. [01-architecture.md](/home/olivier/ll-it-sc/projets/kore/technical/foundation/01-architecture.md) §6) conservée : une seule base Cloud SQL, une entrée de cache préfixée par tenant.
- Routage tenant : sous-domaine → même service Cloud Run (résolution du tenant depuis le JWT/host). Décision URL ouverte (spec §4.5, §17).

## 8. Observabilité

- **Cloud Logging** (logs JSON structurés), **Cloud Monitoring** (métriques Cloud Run/SQL/Redis), **Error Reporting**.
- Endpoints `GET /health` (liveness) et `GET /ready` (readiness : DB + Redis joignables).

## 9. Environnements

| Env | Compute | Cloud SQL | Redis |
| --- | --- | --- | --- |
| Local | Docker Compose | conteneur `postgres:16` | conteneur `redis:7` |
| Staging | Cloud Run | Cloud SQL (petite instance) | Memorystore (partagé) |
| Prod | Cloud Run (min=1, HA) | Cloud SQL HA | Memorystore (partagé) |

La **parité dev/prod** est assurée par des variables d'environnement identiques ; seuls les endpoints changent.

## 10. Definition of Done (fondation GCP)

- [ ] Cible Cloud Run (API + frontend) documentée et paramétrée.
- [ ] Connexion Cloud SQL (Auth Proxy/IP privée) via VPC connector spécifiée.
- [ ] Accès Memorystore privé + contraintes d'instance partagée actées.
- [ ] Secrets centralisés dans Secret Manager (aucun secret committé).
- [ ] Pipeline Cloud Build (test → image → migrations → deploy → smoke) défini.
- [ ] Migrations exécutées en job avant bascule de trafic (pas au boot en prod).
