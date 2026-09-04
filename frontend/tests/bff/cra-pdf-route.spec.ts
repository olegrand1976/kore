import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = join(__dirname, '../..')

function readRoute(rel: string): string {
  return readFileSync(join(root, 'server/api', rel), 'utf8')
}

const PDF_ROUTE = 'cra/timesheets/[id]/pdf.post.ts'

describe('BFF CRA PDF route', () => {
  it('relays the PDF bytes as a Buffer, never a bare ArrayBuffer', () => {
    const src = readRoute(PDF_ROUTE)
    expect(src).toContain('apiAuthHeaders(event)')
    expect(src).toContain('/pdf')
    expect(src).toContain("responseType: 'arrayBuffer'")
    expect(src).toContain('Buffer.from(response._data')
    // `return response._data` seul renvoyait un ArrayBuffer nu : h3 le sérialisait
    // en JSON et le fichier téléchargé faisait 2 octets (`{}`).
    expect(src).not.toMatch(/return\s+response\._data\s*$/m)
  })

  it('forwards the upstream content-type and content-disposition', () => {
    const src = readRoute(PDF_ROUTE)
    expect(src).toContain("setHeader(event, 'content-type'")
    expect(src).toContain("setHeader(event, 'content-disposition'")
  })

  it('documents why a bare ArrayBuffer breaks h3 binary responses', () => {
    // h3 (`handleHandlerResponse`) n'envoie en binaire que si `val.buffer` existe
    // ou si `val.arrayBuffer` est une fonction ; sinon il tombe sur JSON.stringify.
    const raw = new ArrayBuffer(8)
    expect((raw as unknown as { buffer?: unknown }).buffer).toBeUndefined()
    expect(typeof (raw as unknown as { arrayBuffer?: unknown }).arrayBuffer).toBe('undefined')
    expect(JSON.stringify(raw)).toBe('{}')

    const wrapped = Buffer.from(raw)
    expect(wrapped.buffer).toBeTruthy()
    expect(wrapped.length).toBe(8)
  })

  it('preserves the PDF signature through the Buffer conversion', () => {
    const pdf = new TextEncoder().encode('%PDF-1.4\n%âãÏÓ\n')
    const wrapped = Buffer.from(pdf.buffer.slice(0) as ArrayBuffer)
    expect(wrapped.subarray(0, 5).toString('latin1')).toBe('%PDF-')
    expect(wrapped.length).toBe(pdf.byteLength)
  })
})
