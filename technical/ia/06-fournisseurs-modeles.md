# 06 — Fournisseurs et modèles LLM

> Adapters LLM, embeddings, coûts.
> Référence : [01-architecture-module-ai.md](01-architecture-module-ai.md).

## 1. Providers supportés

| Provider | Env | Usage | Données |
| --- | --- | --- | --- |
| **stub** | `AI_LLM_PROVIDER=stub` | Dev, CI, démo sans clé API | Local uniquement |
| **gemini** | `AI_LLM_PROVIDER=gemini` | Production cloud (Google Gemini) | Transfert selon région Google |
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

**Runtime actuel** : un seul modèle Gemini global (`org.platform_settings.gemini_model`, défaut `GEMINI_MODEL` / `gemini-3.6-flash`). Pas de routage par capability.

Reco produit (cible future ou choix admin manuel) :

| Capability | Modèle recommandé | Tokens estimés |
| --- | --- | --- |
| analysis_draft | gemini-3.6-flash / stub | ~1500 |
| classify | gemini-3.5-flash-lite / stub | ~200 |
| dashboard.briefing | gemini-3.6-flash / stub | ~500 |
| publicsite.chatbot | gemini-3.5-flash-lite / stub | ~800 |

Modèles Gemini courants (admin plateforme / `GEMINI_MODEL`) :

| Modèle | Usage |
| --- | --- |
| `gemini-3.6-flash` | Défaut prod — agentique / multimodal, moins de jetons et moins cher que 3.5 Flash |
| `gemini-3.5-flash-lite` | Haut débit, latence et coût minimaux (à sélectionner manuellement côté admin) |
| `gemini-3.5-flash` | Encore supporté (saisie libre / rétrocompat) |

Rate-limit BFF public chat : 20 req/min/IP (roadmap Redis).

## 6. Cache

Redis cache-aside pour :

- Briefing dashboard (TTL 5 min, clé `kore:{tenant}:ai:briefing:{userId}`)
- Classify demand (TTL 1h par sujet hash)

## 7. Definition of Done

- [ ] Stub provider implémenté et testé
- [ ] Interface `LLMProvider` documentée
- [ ] Config env documentée dans README dev
