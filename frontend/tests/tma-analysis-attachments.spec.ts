import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const panelSrc = readFileSync(
  join(__dirname, '../components/requests/RequestAttachmentsPanel.vue'),
  'utf8'
)
const tmaDetailSrc = readFileSync(join(__dirname, '../pages/tma/[id].vue'), 'utf8')

describe('RequestAttachmentsPanel', () => {
  it('supports embedded mode with unique upload input id', () => {
    expect(panelSrc).toContain('embedded?: boolean')
    expect(panelSrc).toContain('inputId?: string')
    expect(panelSrc).toContain('uploadInputId')
    expect(panelSrc).toContain('request-attachments--embedded')
    expect(panelSrc).toContain(':deep(.app-btn)')
    expect(panelSrc).toContain('var(--kore-error)')
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
})
