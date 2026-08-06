import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = join(__dirname, '../..')

function readRoute(rel: string): string {
  return readFileSync(join(root, 'server/api', rel), 'utf8')
}

describe('BFF invoicing routes', () => {
  it('proxies invoice creation to Go POST /api/v1/invoices', () => {
    const src = readRoute('invoices/index.post.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/invoices')
    expect(src).toMatch(/method:\s*['"]POST['"]/)
  })

  it('proxies bulk CRA invoicing to Go POST /prestations/create-invoices', () => {
    const src = readRoute('prestations/create-invoices.post.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/prestations/create-invoices')
    expect(src).toMatch(/method:\s*['"]POST['"]/)
  })
})
