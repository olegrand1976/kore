import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = join(__dirname, '../..')

function readRoute(rel: string): string {
  return readFileSync(join(root, 'server/api', rel), 'utf8')
}

describe('BFF branding routes', () => {
  it('proxies tenant logo bytes with proxyRequest + arrayBuffer', () => {
    const src = readRoute('org/branding/logo/[tenantId].get.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/branding/logo/')
    expect(src).toContain('proxyRequest')
    expect(src).toContain("responseType: 'arrayBuffer'")
  })

  it('proxies societe branding PUT multipart with proxyRequest (raw body)', () => {
    const src = readRoute('org/societes/[id]/branding.put.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/societes/')
    expect(src).toContain('/branding')
    expect(src).toContain('proxyRequest')
    expect(src).not.toContain('readMultipartFormData')
    expect(src).not.toContain('new Blob')
    expect(src).not.toContain('createError')
  })
})
