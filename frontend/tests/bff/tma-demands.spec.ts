import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const createSrc = readFileSync(join(__dirname, '../../server/api/tma/demands.post.ts'), 'utf8')
const deleteSrc = readFileSync(join(__dirname, '../../server/api/tma/demands/[id].delete.ts'), 'utf8')

describe('BFF TMA demand create', () => {
  it('proxies POST /api/v1/demands with auth headers', () => {
    expect(createSrc).toContain('apiAuthHeaders(event)')
    expect(createSrc).toContain('/api/v1/demands')
    expect(createSrc).toMatch(/method:\s*['"]POST['"]/)
  })

  it('forwards Go API errors instead of Nitro Server Error', () => {
    expect(createSrc).toContain('createError')
    expect(createSrc).toContain('err.data')
    expect(createSrc).toContain('statusMessage')
  })
})

describe('BFF TMA demand delete', () => {
  it('proxies DELETE /api/v1/demands/:id with auth headers', () => {
    expect(deleteSrc).toContain('apiAuthHeaders(event)')
    expect(deleteSrc).toContain('/api/v1/demands/')
    expect(deleteSrc).toMatch(/method:\s*['"]DELETE['"]/)
  })

  it('forwards Go API errors instead of Nitro Server Error', () => {
    expect(deleteSrc).toContain('createError')
    expect(deleteSrc).toContain('err.data')
    expect(deleteSrc).toContain('statusMessage')
  })
})
