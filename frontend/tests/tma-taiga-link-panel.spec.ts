import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const bffSrc = readFileSync(
  join(__dirname, '../server/api/integrations/taiga/links/by-demand/[id].get.ts'),
  'utf8'
)
const panelSrc = readFileSync(join(__dirname, '../components/tma/TaigaLinkPanel.vue'), 'utf8')
const tmaDetailSrc = readFileSync(join(__dirname, '../pages/tma/[id].vue'), 'utf8')

describe('BFF Taiga link by demand', () => {
  it('proxies GET with auth headers', () => {
    expect(bffSrc).toContain('apiAuthHeaders(event)')
    expect(bffSrc).toContain('/api/v1/integrations/taiga/links/by-demand/')
  })
})

describe('TaigaLinkPanel', () => {
  it('loads link from BFF and handles missing link', () => {
    expect(panelSrc).toContain('/api/integrations/taiga/links/by-demand/')
    expect(panelSrc).toContain('tma.taiga_not_linked')
    expect(panelSrc).toContain('tma.taiga_ref')
  })

  it('reloads when demandId changes', () => {
    expect(panelSrc).toContain('watch(')
    expect(panelSrc).toContain('props.demandId')
    expect(panelSrc).toContain('immediate: true')
    expect(panelSrc).toContain('loadGeneration')
  })

  it('shows linked state when ref or url is present', () => {
    expect(panelSrc).toContain('hasLink')
    expect(panelSrc).toContain('externalUrl.value')
  })
})

describe('TMA detail Taiga panel', () => {
  it('embeds TaigaLinkPanel for tma readers', () => {
    expect(tmaDetailSrc).toContain('TaigaLinkPanel')
    expect(tmaDetailSrc).toContain("can('tma', 'L')")
  })
})
