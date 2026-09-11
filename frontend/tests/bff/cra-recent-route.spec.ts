import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = join(__dirname, '../..')

describe('BFF CRA timesheets/recent route', () => {
  it('forwards year, month, userId and limit query params', () => {
    const src = readFileSync(join(root, 'server/api/cra/timesheets/recent.get.ts'), 'utf8')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain("['limit', 'year', 'month', 'userId']")
    expect(src).toContain('/api/v1/timesheets/recent')
    expect(src).toContain('params.set(key, String')
  })
})
