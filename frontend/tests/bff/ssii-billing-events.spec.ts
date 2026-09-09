import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = join(__dirname, '../..')

function readRoute(rel: string): string {
  return readFileSync(join(root, 'server/api', rel), 'utf8')
}

describe('BFF SSII mission billing-events routes', () => {
  it('proxies GET list with auth headers', () => {
    const src = readRoute('ssii/missions/[id]/billing-events.get.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/billing-events')
  })

  it('proxies POST note with body', () => {
    const src = readRoute('ssii/missions/[id]/billing-events.post.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('readBody')
    expect(src).toContain('method: \'POST\'')
    expect(src).toContain('/billing-events')
  })
})
