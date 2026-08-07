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

  it('proxies CRA invoice preview to Go POST /prestations/preview-invoices', () => {
    const src = readRoute('prestations/preview-invoices.post.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/api/v1/prestations/preview-invoices')
    expect(src).toMatch(/method:\s*['"]POST['"]/)
  })

  it('proxies proforma emission to Go POST /invoices/{id}/emit-proforma', () => {
    const src = readRoute('invoices/[id]/emit-proforma.post.ts')
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('emit-proforma')
    expect(src).toContain('x-public-base-url')
    expect(src).toMatch(/method:\s*['"]POST['"]/)
  })

  it('proxies public proforma preview and validate without auth headers', () => {
    const getSrc = readRoute('public/proforma/[token].get.ts')
    const postSrc = readRoute('public/proforma/[token]/validate.post.ts')
    expect(getSrc).toContain('/api/v1/public/proforma/')
    expect(getSrc).not.toContain('apiAuthHeaders')
    expect(postSrc).toContain('/validate')
    expect(postSrc).toMatch(/method:\s*['"]POST['"]/)
  })
})
