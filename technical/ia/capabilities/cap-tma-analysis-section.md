# cap-tma-analysis-section — Section dossier analyse TMA (RAG)

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `tma.analysis_section` |
| Module | 05 TMA |
| Vague | 1 — P0 |
| Statut | livré |

## Finalité métier

Générer **une section** du dossier d'analyse (fonctionnel, technique, risques, scénario de test) à partir d'une consigne utilisateur, avec option **RAG** sur les pièces jointes indexées de la demande (PDF, texte, CSV…).

Persona : développeur TMA / analyste.

## Classification IA Act

- **Risque** : minimal / limité (contenu généré Art. 50)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous RG-TMA-01

- Ne modifie pas le statut demande ni la gate chef utilisateur
- Texte **non enregistré** tant que l'utilisateur n'a pas cliqué « Enregistrer » sur le dossier
- RAG limité au tenant + `demand_id` (chunks `ai.document_chunks`)
- Validation humaine obligatoire avant résolution workflow

## Entrées / sorties

**Entrée** : `demandId`, `section` (`functional` \| `technical` \| `risks` \| `testScenario`), `prompt`, `useRAG`, `subject` (optionnel)  
**Sortie** : `{ text, requestId, sources? }` — `sources` = métadonnées chunks RAG (nom fichier, index)

## Indexation pièces jointes

- Hook org → module `ai` : indexation asynchrone à l'upload (`IndexRequestAttachment`)
- Suppression PJ → `RemoveRequestAttachmentChunks`
- Embeddings : Gemini (`text-embedding-004`) ou stub déterministe (CI)

## Ancrage code

- [`frontend/components/tma/AnalysisEditor.vue`](../../../frontend/components/tma/AnalysisEditor.vue)
- [`frontend/pages/tma/[id].vue`](../../../frontend/pages/tma/[id].vue)
- [`frontend/components/requests/RequestAttachmentsPanel.vue`](../../../frontend/components/requests/RequestAttachmentsPanel.vue)
- `internal/modules/ai/app/analysis_section.go` — `SuggestAnalysisSection`
- `internal/modules/ai/app/rag.go` — retrieval pgvector

## API

- BFF : `POST /api/ai/tma/analysis-section`
- Go : `POST /api/v1/ai/tma/analysis-section`

## UX

- Par section : champ consigne + case « Utiliser les pièces jointes indexées (RAG) » (activée si PJ indexables présentes)
- Bouton « Générer la section » + `AppAiBadge` + disclaimer
- i18n `ai.section_*`, `ai.use_rag`, `ai.rag_no_docs`

## DoD

- [x] Route BFF + handler Go + OpenAPI
- [x] Migration `ai.document_chunks` + capability seed
- [x] Journalisation `ai_request_log` + explicabilité sources RAG
- [x] Badge UI + i18n
- [x] Tests unitaires chunking / section / stub embed ; intégration postgres chunks
