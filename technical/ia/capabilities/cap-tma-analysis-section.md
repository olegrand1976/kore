# cap-tma-analysis-section — Section dossier analyse TMA (contexte documents)

## Identifiant

| Champ | Valeur |
| --- | --- |
| Code | `tma.analysis_section` |
| Module | 05 TMA |
| Vague | 1 — P0 |
| Statut | livré |

## Finalité métier

Générer **une section** du dossier d'analyse (fonctionnel, technique, risques, scénario de test) à partir d'une consigne utilisateur, avec option d'utiliser les **pièces jointes indexées** de la demande (PDF texte, TXT, MD, CSV…).

Persona : développeur TMA / analyste.

## Classification IA Act

- **Risque** : minimal / limité (contenu généré Art. 50)
- **Annexe III** : Non
- **Art. 6(3)** : N/A

## Garde-fous RG-TMA-01

- Ne modifie pas le statut demande ni la gate chef utilisateur
- Texte **non enregistré** tant que l'utilisateur n'a pas cliqué « Enregistrer » sur le dossier
- Retrieval limité au tenant + `demand_id` (chunks `ai.document_chunks`)
- Validation humaine obligatoire avant résolution workflow

## Entrées / sorties

**Entrée** : `demandId`, `section` (`functional` \| `technical` \| `risks` \| `testScenario`), `prompt`, `useRAG`, `subject` (optionnel)  
**Sortie** : `{ text, requestId, usedDocuments, sources? }` — `sources` = métadonnées chunks (nom fichier, index)

**Statut indexation** : `GET /ai/tma/document-context?demandId=` → `{ indexedChunkCount, hasIndexedDocuments }`

## Indexation pièces jointes

- Hook org → module `ai` : indexation à l'upload (`IndexRequestAttachment`)
- Suppression PJ → `RemoveRequestAttachmentChunks`
- Embeddings : Gemini (`text-embedding-004`) ou stub déterministe (CI)
- La case « s'appuyer sur les PJ » côté UI s'active uniquement si `hasIndexedDocuments` (chunks réels), pas seulement sur l'extension fichier

## Ancrage code

- [`frontend/components/tma/AnalysisEditor.vue`](../../../frontend/components/tma/AnalysisEditor.vue)
- [`frontend/pages/tma/[id].vue`](../../../frontend/pages/tma/[id].vue)
- [`frontend/components/requests/RequestAttachmentsPanel.vue`](../../../frontend/components/requests/RequestAttachmentsPanel.vue)
- `internal/modules/ai/app/analysis_section.go` — `SuggestAnalysisSection`
- `internal/modules/ai/app/rag.go` — retrieval pgvector + `DemandDocumentContext`

## API

- BFF : `POST /api/ai/tma/analysis-section`, `GET /api/ai/tma/document-context`
- Go : `POST /api/v1/ai/tma/analysis-section`, `GET /api/v1/ai/tma/document-context`

## UX

- Par section : champ consigne + case « S'appuyer sur les pièces jointes du dossier » (activée si documents **indexés**)
- Bouton « Générer la section » + `AppAiBadge` + disclaimer
- Liste des sources utilisées ; message si le contexte documents a été demandé mais non appliqué
- i18n `ai.section_*`, `ai.use_attachments`, `ai.attachments_*`, `ai.section_sources`

## DoD

- [x] Route BFF + handler Go + OpenAPI
- [x] Migration `ai.document_chunks` + capability seed
- [x] Journalisation `ai_request_log` + explicabilité sources
- [x] Badge UI + i18n (sans jargon technique côté utilisateur)
- [x] Statut d'indexation réel (`document-context`)
- [x] Tests unitaires chunking / section / stub embed ; intégration postgres chunks
