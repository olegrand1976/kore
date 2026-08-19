import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const startSrc = readFileSync(
  join(__dirname, '../../server/api/project/applications/[appId]/sprints/[sprintId]/start.post.ts'),
  'utf8'
)
const closeSrc = readFileSync(
  join(__dirname, '../../server/api/project/applications/[appId]/sprints/[sprintId]/close.post.ts'),
  'utf8'
)
const planSrc = readFileSync(
  join(__dirname, '../../server/api/project/applications/[appId]/sprints/[sprintId]/plan.post.ts'),
  'utf8'
)

describe('BFF project sprint actions', () => {
  it('proxies sprint start to Go API', () => {
    expect(startSrc).toContain('apiAuthHeaders(event)')
    expect(startSrc).toContain('/sprints/${sprintId}/start')
    expect(startSrc).toMatch(/method:\s*['"]POST['"]/)
  })

  it('proxies sprint close to Go API', () => {
    expect(closeSrc).toContain('/sprints/${sprintId}/close')
    expect(closeSrc).toMatch(/method:\s*['"]POST['"]/)
  })

  it('proxies sprint plan to Go API', () => {
    expect(planSrc).toContain('/sprints/${sprintId}/plan')
    expect(planSrc).toMatch(/method:\s*['"]POST['"]/)
  })
})
