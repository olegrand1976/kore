# 00 — Conformité IA Act européen

> Fondation réglementaire pour toutes les capabilities IA Kore.
> Référence : Règlement (UE) 2024/1689.

## 1. Rôles

| Rôle | Acteur | Obligations |
| --- | --- | --- |
| **Provider** | Kore (éditeur SaaS) | Art. 9–15, 50, 71 — conception, documentation, journalisation |
| **Deployer** | Tenant ESN/DSI | Art. 26 — activation, information travailleurs, usage conforme |

## 2. Calendrier

| Échéance | Obligation |
| --- | --- |
| Fév. 2025 | Pratiques interdites (Art. 5) |
| Août 2026 | Obligations GPAI |
| Déc. 2027 | Systèmes haut risque Annexe III : conformité, CE, base UE |

## 3. Classification des risques

| Niveau | Exemples Kore | Obligations |
| --- | --- | --- |
| **Interdit** | Émotions au travail, scoring social, décisions RH auto | Ne pas implémenter |
| **Haut risque** | Matching M08, monitoring performance, aide décision congés | Chapitre III + Art. 6(3) ou exclusion |
| **Limité** | Chatbot, contenu généré TMA | Art. 50 — disclosure utilisateur |
| **Minimal** | Brouillon analyse, doublons TMA, budget | Logging, opt-in |

## 4. Pratiques interdites (Art. 5)

- Reconnaissance des **émotions** en contexte professionnel
- **Scoring social** ou classement de collaborateurs
- Décisions **automatisées** : validation congés, gate TMA, CRA définitif, affectation sans humain
- Inférence traits personnels pour allocation tâches ou évaluation performance
- Entraînement sur données tenant sans base légale

## 5. Obligations provider Kore

| Article | Implémentation |
| --- | --- |
| Art. 9 | Registre capabilities + revue périodique |
| Art. 10 | Isolation tenant, minimisation PII dans prompts |
| Art. 11 | Fiche par capability (`capabilities/cap-*.md`) |
| Art. 12 | Table `ai.ai_request_log` |
| Art. 13 | Notice deployer (`03-governance-deployer.md`) |
| Art. 14 | Pattern accept/modify/reject UI |
| Art. 15 | Tests stub + disclaimer hallucinations |
| Art. 50 | `AppAiBadge` |
| Art. 71 | Enregistrement UE si haut risque confirmé |

## 6. Obligations deployer (facilitation Kore)

| Article | Facilitation |
| --- | --- |
| Art. 26(2) | Notice + checklist activation admin |
| Art. 26(7) | Workflow « informer les travailleurs » |
| Art. 26(11) / Art. 86 | `GET /api/v1/ai/explain/{requestId}` |
| Art. 27 | Template FRIA (secteur public) — roadmap |

## 7. Art. 6(3) — Réclassification

Tout système Annexe III peut être **non haut-risque** si documenté. Kore produit une évaluation écrite par capability ambiguë dans [02-capabilities-registry.md](02-capabilities-registry.md).

## 8. Definition of Done

- [ ] Registre capabilities à jour
- [ ] Aucune capability interdite en production
- [ ] Capabilities haut risque : évaluation Art. 6(3) ou report phase ultérieure
- [ ] Notice deployer FR/EN disponible
