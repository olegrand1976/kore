import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const hubSrc = readFileSync(join(__dirname, '../pages/aide/index.vue'), 'utf8')
const tmaHelpSrc = readFileSync(join(__dirname, '../pages/aide/tma.vue'), 'utf8')

describe('Aide TMA / Taiga', () => {
  it('links Taiga help topic from hub', () => {
    expect(hubSrc).toContain('/aide/tma')
    expect(hubSrc).toContain('help.topics.tma.title')
  })

  it('documents Taiga panel and linking flow', () => {
    expect(tmaHelpSrc).toContain('help.tma.panel_title')
    expect(tmaHelpSrc).toContain('help.tma.link_body')
    expect(tmaHelpSrc).toContain('/tma')
  })
})
