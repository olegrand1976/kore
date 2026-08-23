# Test flux Taiga ↔ Kore — 4 niveaux (staging)

**Date** : 2026-08-23  
**Environnement** : `https://kore.ll-it-sc.be` + Taiga `https://taiga.ll-it-sc.be`  
**Compte test** : `ADM_admin` (tenant seed `00000000-0000-4000-8000-000000000001`)  
**Commit déployé** : `0271c7a` (durcissement édition) + tests smoke

---

## Synthèse

| Niveau | Cible | Résultat staging | Bloquant ops / code |
|--------|-------|------------------|---------------------|
| **1** Utilisateurs | Kore user ↔ Taiga user | **Implémenté** — GET + UI `/admin/integrations` | Consommation webhook à faire |
| **2** Applications | Kore application ↔ Taiga project | **Bloqué ops** — API 503 `taiga not configured` | Exécuter `taiga-setup-gcp.sh` + redeploy |
| **3** Tâches / backlog | US + tasks Taiga ↔ entités Kore | **Non implémenté** | Pas de `kore_entity_type` backlog/work_item |
| **4** Incidents / issues | Demandes TMA ↔ userstory/issue/task | **OK webhook** — fix `external_type` en branche | Pas de type `issue` Kore |

---

## Prérequis vérifiés

| Élément | Statut |
|---------|--------|
| Login `ADM_admin` | OK |
| Instance Taiga HTTP | 200 |
| Secret `kore-taiga-webhook-secret` (GCP) | OK |
| Secret `kore-taiga-base-url` | `https://taiga.ll-it-sc.be` |
| Secret `kore-taiga-service-username/password` | **Absent** (liste GCP : pas de `kore-taiga-service-*`) |
| Webhook sans secret | 401 (attendu) |

---

## Niveau 1 — Utilisateurs

**API testée** :

```http
POST /api/v1/integrations/taiga/user-mappings
Authorization: Bearer <token>
{
  "taigaUserId": 12345,
  "taigaUsername": "adm_admin",
  "koreUserId": "5d88b377-5a2f-4d7b-be81-1cb99ea7e59a",
  "matchMethod": "email"
}
```

| Vérification | Résultat |
|--------------|----------|
| HTTP | **201 Created** |
| Persistance | Ligne `integrations.user_mappings` (provider `taiga`) |
| GET `/integrations/taiga/user-mappings` | **200** — liste tenant (RBAC `integrations:L`) |
| UI admin | Section **Mappings utilisateurs Taiga** sur `/admin/integrations` |
| Consommation webhook / auth | Aucune (phase ultérieure) |

**Écart résiduel** : usage des mappings dans les flux webhook / auth (phase ultérieure).

---

## Niveau 2 — Applications

**API testée** (token admin) :

| Endpoint | HTTP | Corps / note |
|----------|------|--------------|
| `GET /integrations/taiga/projects/unlinked` | **503** | `taiga not configured` |
| `GET /integrations/taiga/links/by-application/{id}` (apps seed) | **404** | Aucune liaison en base |

**Cause 503** : le client sortant Taiga (`ListProjects`) n’est pas câblé sur staging — variables `TAIGA_SERVICE_USERNAME` / `TAIGA_SERVICE_PASSWORD` non injectées (cf. [`deploy-run-args.sh`](../infra/gcp/lib/deploy-run-args.sh)). Secrets MCP Taiga (`taiga-mcp-taiga-username`) existent mais ne sont pas mappés sur l’API Kore.

**Code déployé** : import masse, création/édition avec `taigaProjectId`, BFF `by-application` — **non testable bout-en-bout** tant que le compte service n’est pas provisionné.

**Action ops** :

```bash
# Créer secrets Kore (compte service lecture projets Taiga)
echo -n '<user>' | gcloud secrets versions add kore-taiga-service-username --project=premedica-prod-2025 --data-file=-
echo -n '<pass>' | gcloud secrets versions add kore-taiga-service-password --project=premedica-prod-2025 --data-file=-
# Redéployer API (staging)
```

Puis retester : `/admin/applications` → import ou édition → GET `by-application` → 200.

---

## Niveau 3 — Tâches / backlog (non implémenté)

| Sous-niveau | Attendu métier | État code |
|-------------|----------------|-----------|
| **3a** User stories ↔ backlog projet | Lier une US Taiga à une demande backlog (`/projets`, champs `epic_id` / `sprint_id`) | Aucun endpoint ; pas de `kore_entity_type` autre que `demand` / `application` |
| **3b** Tasks Taiga ↔ entité Kore | Task Taiga → entité dédiée (hors demande TMA) | Webhook force `KoreEntityType: "demand"` ([`taiga_service.go`](../internal/modules/integrations/app/taiga_service.go) L129) |

**Test de contournement** (niveau 4) : webhook `type: "task"` avec UUID demande → enregistre bien un lien **demande**, pas backlog.

**Écarts P0** : modèle `external_links` + webhook pour backlog ; entité Kore pour tasks Taiga.

---

## Niveau 4 — Incidents / issues (demandes TMA)

### Script nominal

```bash
./scripts/taiga-webhook-smoke.sh https://kore.ll-it-sc.be <uuid-demande-tma>
```

| Demande | Webhook | GET `by-demand` |
|---------|---------|-----------------|
| `7bfd58d9-…` (userstory smoke) | 200 | 200 — `ExternalType=userstory`, `ExternalRef=42`, URL `…/us/42` |
| `b8ff6e62-…` (userstory smoke) | 200 | 200 — lien créé |
| `4f38e3ef-…` (issue puis task) | 200 / 200 | voir ci-dessous |

### Variantes `issue` et `task`

Webhook `type: "issue"` → **OK** :

- `ExternalType`: `issue`
- `KoreEntityType`: `demand`
- `ExternalRef`: 15

Webhook `type: "task"` (même demande, upsert) → **OK** après fix SQL :

- `ExternalType`: `task`
- `ExternalID`: `77001`, `ExternalURL`: `…/task/9`, `ExternalRef`: 9

### UI

- API confirme les liens pour les demandes testées.
- Panneau **Lien Taiga** sur `/tma/{id}` ([`TaigaLinkPanel.vue`](../frontend/components/tma/TaigaLinkPanel.vue)) : non vérifié navigateur (auth manuelle requise) ; données API suffisantes pour valider le flux inbound.

### Limites métier

- Kore TMA : type domaine `incident` uniquement — pas de distinction incident / issue côté Kore.
- Unicité : 1 lien Taiga par demande (`external_links_kore_entity_unique`).

---

## Écarts priorisés (roadmap)

| Priorité | Écart | Action suggérée |
|----------|-------|-----------------|
| **P0 ops** | Niveau 2 inutilisable staging | `./scripts/taiga-setup-gcp.sh` (copie MCP → `kore-taiga-service-*`) + redeploy |
| **P2** | Niveau 3 backlog + tasks | Spec + migrations `external_links` + webhook multi-entités |
| **P2** | Type `issue` TMA Kore | Évolution domaine TMA ou mapping explicite incident/issue |
| **P3** | Sync sortant (création US depuis Kore) | Phase 3 spec §7.2.5 |

---

## Commandes de rejeu

```bash
# Niveau 4 — webhook + lien
./scripts/taiga-webhook-smoke.sh https://kore.ll-it-sc.be <uuid-demande>

# Niveau 2 — après fix secrets service
curl -H "Authorization: Bearer $TOKEN" \
  https://kore.ll-it-sc.be/api/v1/integrations/taiga/projects/unlinked

# Niveau 1 — mapping (intégrations:E)
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"taigaUserId":12345,"taigaUsername":"x","koreUserId":"<uuid>","matchMethod":"email"}' \
  https://kore.ll-it-sc.be/api/v1/integrations/taiga/user-mappings
```

---

## Références

- Runbook : [`technical/modules/17-integrations-hub.md`](../technical/modules/17-integrations-hub.md) §13
- Spec : [`SPECIFICATION_FONCTIONNELLE.md`](SPECIFICATION_FONCTIONNELLE.md) §7.2.5
