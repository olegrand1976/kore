import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = join(__dirname, '../..')

function readRoute(rel: string): string {
  return readFileSync(join(root, 'server/api', rel), 'utf8')
}

describe('BFF ssii mission application routes', () => {
  it('proxies mission applications update to Go PUT /missions/{id}/applications', () => {
    const src = readRoute('ssii/missions/[id]/applications.put.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/missions/')
    expect(src).toContain('/applications')
    expect(src).toMatch(/method:\s*['"]PUT['"]/)
  })

  it('proxies mission create body to Go POST /missions', () => {
    const src = readRoute('ssii/missions.post.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/missions')
    expect(src).toMatch(/method:\s*['"]POST['"]/)
    expect(src).toContain('readBody(event)')
  })

  it('proxies mission detail get', () => {
    const src = readRoute('ssii/missions/[id].get.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/missions/')
  })
})
