import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const panelSrc = readFileSync(
  join(__dirname, '../components/requests/RequestAttachmentsPanel.vue'),
  'utf8'
)
const editorSrc = readFileSync(join(__dirname, '../components/tma/AnalysisEditor.vue'), 'utf8')
const tmaDetailSrc = readFileSync(join(__dirname, '../pages/tma/[id].vue'), 'utf8')
const useAiSrc = readFileSync(join(__dirname, '../composables/useAi.ts'), 'utf8')
const bffDocCtx = readFileSync(
  join(__dirname, '../server/api/ai/tma/document-context.get.ts'),
  'utf8'
)
const frLocale = readFileSync(join(__dirname, '../locales/fr.json'), 'utf8')
const enLocale = readFileSync(join(__dirname, '../locales/en.json'), 'utf8')

describe('RequestAttachmentsPanel', () => {
  it('supports embedded mode with unique upload input id', () => {
    expect(panelSrc).toContain('embedded?: boolean')
    expect(panelSrc).toContain('inputId?: string')
    expect(panelSrc).toContain('uploadInputId')
    expect(panelSrc).toContain('request-attachments--embedded')
    expect(panelSrc).toContain(':deep(.app-btn)')
    expect(panelSrc).toContain('var(--kore-error)')
  })

  it('emits indexable docs from document-context status when TMA', () => {
    expect(panelSrc).toContain('fetchDocumentContext')
    expect(panelSrc).toContain('indexableDocsChanged')
    expect(panelSrc).toContain('emitIndexableDocs')
    expect(panelSrc).toContain("props.resource === 'tma'")
  })

  it('allows deleting attachments', () => {
    expect(panelSrc).toContain('onDelete')
    expect(panelSrc).toContain('request-attachments__delete')
    expect(panelSrc).toContain('common.delete')
  })

  it('includes preview UI and i18n keys', () => {
    expect(panelSrc).toContain('attachments_preview')
    expect(panelSrc).toContain('previewKind')
    expect(panelSrc).toContain('previewSeq')
    expect(frLocale).toContain('"attachments_preview"')
    expect(frLocale).toContain('"attachments_preview_unsupported"')
  })
})

describe('TMA detail analysis attachments', () => {
  it('embeds a single attachments panel inside the analysis dossier', () => {
    expect(tmaDetailSrc).toContain('tma.analysis_attachments')
    expect(tmaDetailSrc).toContain('embedded')
    expect(tmaDetailSrc).toMatch(/RequestAttachmentsPanel[\s\S]*analysis_attachments/)
    const panelCount = (tmaDetailSrc.match(/<RequestAttachmentsPanel/g) ?? []).length
    expect(panelCount).toBe(1)
  })

  it('wires indexable attachments to AnalysisEditor document context toggle', () => {
    expect(tmaDetailSrc).toContain('hasIndexableDocs')
    expect(tmaDetailSrc).toContain('indexable-docs-changed')
    expect(tmaDetailSrc).toContain(':has-indexable-docs')
  })
})

describe('AnalysisEditor section AI', () => {
  it('exposes per-section prompt, document context and generate controls', () => {
    expect(editorSrc).toContain('hasIndexableDocs')
    expect(editorSrc).toContain('generateAnalysisSection')
    expect(editorSrc).toContain('sectionUseAttachments')
    expect(editorSrc).not.toContain('sectionUseRAG')
    expect(editorSrc).toContain('ai.section_prompt')
    expect(editorSrc).toContain('ai.use_attachments')
    expect(editorSrc).toContain('ai.section_generate')
    expect(editorSrc).toContain('sectionSources')
    expect(editorSrc).toContain('attachments_not_applied')
    expect(frLocale).toContain('"section_prompt"')
    expect(frLocale).toContain('"use_attachments"')
    expect(frLocale).toContain('"attachments_unavailable_for_ai"')
    expect(frLocale).toContain('"section_sources"')
    expect(frLocale).not.toContain('(RAG)')
    expect(enLocale).toContain('"use_attachments"')
    expect(enLocale).toContain('"attachments_unavailable_for_ai"')
    expect(enLocale).toContain('Use the request attachments as context')
    expect(enLocale).not.toContain('(RAG)')
    expect(enLocale).not.toContain('dossier attachments')
  })

  it('uses document-context BFF and unwraps AI envelopes', () => {
    expect(useAiSrc).toContain('fetchDocumentContext')
    expect(useAiSrc).toContain('document-context')
    expect(useAiSrc).toContain('unwrapData')
    expect(useAiSrc).toContain('usedDocuments')
    expect(bffDocCtx).toContain('/api/v1/ai/tma/document-context')
  })
})
