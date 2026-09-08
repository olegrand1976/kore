import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = join(__dirname, '../..')

function readRoute(rel: string): string {
  return readFileSync(join(root, 'server/api', rel), 'utf8')
}

describe('BFF CRA validate route', () => {
  it('forwards force body and preserves upstream error data', () => {
    const src = readRoute('cra/timesheets/[id]/validate.post.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/validate')
    expect(src).toContain('readBody')
    expect(src).toContain('force')
    expect(src).toContain('createError')
    expect(src).toContain('data: err.data')
  })
})
