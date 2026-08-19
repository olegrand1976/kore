import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const src = readFileSync(join(__dirname, '../../server/api/tma/demands.post.ts'), 'utf8')

describe('BFF TMA demand create', () => {
  it('proxies POST /api/v1/demands with auth headers', () => {
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/demands')
    expect(src).toMatch(/method:\s*['"]POST['"]/)
  })

  it('forwards Go API errors instead of Nitro Server Error', () => {
    expect(src).toContain('createError')
    expect(src).toContain('err.data')
    expect(src).toContain('statusMessage')
  })
})
