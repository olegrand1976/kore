# 06 — Fournisseurs et modèles LLM

> Adapters LLM, embeddings, coûts.
> Référence : [01-architecture-module-ai.md](01-architecture-module-ai.md).

## 1. Providers supportés

| Provider | Env | Usage | Données |
| --- | --- | --- | --- |
| **stub** | `AI_LLM_PROVIDER=stub` | Dev, CI, démo sans clé API | Local uniquement |
| **openai** | `AI_LLM_PROVIDER=openai` | Production cloud | Transfert UE/US selon config |
| **ollama** | `AI_LLM_PROVIDER=ollama` | Souveraineté, on-prem | Local |

Défaut : **stub** — génération heuristique structurée.

## 2. Adapter stub

- Pas d'appel réseau
- Templates par capability (sujet TMA → sections analyse)
- Classification par mots-clés (bug, régression, évolution)
- Similarité : recherche textuelle LIKE sur sujets TMA résolus (sans pgvector en V1)

## 3. Embeddings (roadmap)

- Extension PostgreSQL `pgvector`
- Table `ai.demand_embeddings` (tenant_id, demand_id, vector)
- Isolation stricte par tenant_id dans requêtes

## 4. Choix par tenant

Colonne `tenant_ai_settings.llm_provider` override env global (admin).

Priorité : tenant > env > stub.

## 5. Coûts et limites

| Capability | Modèle recommandé | Tokens estimés |
| --- | --- | --- |
| analysis_draft | gpt-4o-mini / stub | ~1500 |
| classify | petit / stub | ~200 |
| dashboard.briefing | gpt-4o-mini / stub | ~500 |
| publicsite.chatbot | gpt-4o-mini / stub | ~800 |

Rate-limit BFF public chat : 20 req/min/IP (roadmap Redis).

## 6. Cache

Redis cache-aside pour :

- Briefing dashboard (TTL 5 min, clé `kore:{tenant}:ai:briefing:{userId}`)
- Classify demand (TTL 1h par sujet hash)

## 7. Definition of Done

- [ ] Stub provider implémenté et testé
- [ ] Interface `LLMProvider` documentée
- [ ] Config env documentée dans README dev
