import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const getSrc = readFileSync(
  join(__dirname, '../../server/api/project/applications/[appId]/kanban-config.get.ts'),
  'utf8'
)
const putSrc = readFileSync(
  join(__dirname, '../../server/api/project/applications/[appId]/kanban-config.put.ts'),
  'utf8'
)

describe('BFF project kanban-config', () => {
  it('proxies kanban config GET to Go API', () => {
    expect(getSrc).toContain('apiAuthHeaders(event)')
    expect(getSrc).toContain('/kanban-config')
  })

  it('proxies kanban config PUT to Go API', () => {
    expect(putSrc).toContain('apiAuthHeaders(event)')
    expect(putSrc).toContain('/kanban-config')
    expect(putSrc).toMatch(/method:\s*['"]PUT['"]/)
  })
})
