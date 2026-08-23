import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const mergeSrc = readFileSync(join(__dirname, '../../server/api/org/applications/merge.post.ts'), 'utf8')
const taigaAppsSrc = readFileSync(
  join(__dirname, '../../server/api/integrations/taiga/links/applications.get.ts'),
  'utf8'
)

describe('BFF applications merge', () => {
  it('proxies POST merge with auth headers', () => {
    expect(mergeSrc).toContain('apiAuthHeaders(event)')
    expect(mergeSrc).toContain('/api/v1/applications/merge')
    expect(mergeSrc).toContain('POST')
  })

  it('proxies GET taiga linked application ids', () => {
    expect(taigaAppsSrc).toContain('apiAuthHeaders(event)')
    expect(taigaAppsSrc).toContain('/api/v1/integrations/taiga/links/applications')
  })
})
