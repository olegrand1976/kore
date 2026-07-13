# Assistance IA — Spécifications techniques Kore

> Dossier transversal documentant l'intégration IA sur Kore (PSA/ESN). Complète [`technical/README.md`](../README.md) et la roadmap post-MVP.
>
> Références : Règlement UE 2024/1689 (IA Act), [`documentation/SPECIFICATION_FONCTIONNELLE.md`](../../documentation/SPECIFICATION_FONCTIONNELLE.md).

## Objectif

Centraliser la **stratégie**, la **conformité IA Act**, l'**architecture hexagonale** du module `internal/modules/ai/` et une **fiche par capability** (point d'implémentation IA).

## Phasing

| Vague | Horizon | Livrables doc + code |
| --- | --- | --- |
| **0** | Prérequis | Fiches socle 00–06, registre capabilities, module `ai/` squelette |
| **1** | 4–6 sem. | TMA : brouillon analyse, classification, doublons |
| **2** | 6–8 sem. | CRA prefill, anomalies factuelles, budget estimation, dashboard briefing |
| **3** | 8–12 sem. | Congés manager, workflow explain, chatbot publicsite |
| **4** | Phase 2+ | Facturation M09, ETT M10, integrations M17 |

## Fiches socle

| Fiche | Contenu |
| --- | --- |
| [00-ai-act-conformite.md](00-ai-act-conformite.md) | Cadre IA Act : rôles provider/deployer, calendrier, interdits |
| [01-architecture-module-ai.md](01-architecture-module-ai.md) | Module hexagonal `ai/`, ports, BFF, entitlements |
| [02-capabilities-registry.md](02-capabilities-registry.md) | Registre officiel : code, risque, vague, Art. 6(3) |
| [03-governance-deployer.md](03-governance-deployer.md) | Opt-in tenant, notice, information travailleurs |
| [04-journalisation-explicabilite.md](04-journalisation-explicabilite.md) | `ai_request_log`, endpoint explain, rétention |
| [05-ui-patterns.md](05-ui-patterns.md) | `AppAiBadge`, accept/reject, i18n `ai.*` |
| [06-fournisseurs-modeles.md](06-fournisseurs-modeles.md) | LLM stub / cloud / Ollama, pgvector |

## Capabilities (points d'implémentation)

Index : [capabilities/README.md](capabilities/README.md)

## Gate validation documentation

Une vague doc est **complète** quand :

- [ ] Toutes les capabilities de la vague ont une fiche `cap-*.md`
- [ ] [02-capabilities-registry.md](02-capabilities-registry.md) est à jour
- [ ] [technical/README.md](../README.md) référence ce dossier
- [ ] Capabilities haut risque : évaluation Art. 6(3) documentée
- [ ] Contrats API BFF et ports Go décrits avant implémentation

## Principes impératifs

1. **Assistive, jamais autoritaire** — l'humain valide (Art. 14 IA Act)
2. **Pas de LLM côté browser** — BFF Nitro uniquement
3. **Tenant-scopé** — opt-in admin, journalisation Art. 12
4. **Transparence UI** — badge « généré par IA » (Art. 50)
5. **RG-CRA-01** — prefill IA ne remplace jamais `OriginManual`
