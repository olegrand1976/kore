import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const unlinkedSrc = readFileSync(
  join(__dirname, '../../server/api/integrations/taiga/projects/unlinked.get.ts'),
  'utf8'
)
const byApplicationSrc = readFileSync(
  join(__dirname, '../../server/api/integrations/taiga/links/by-application/[id].get.ts'),
  'utf8'
)
const importSrc = readFileSync(
  join(__dirname, '../../server/api/integrations/taiga/applications/import.post.ts'),
  'utf8'
)
const appsPageSrc = readFileSync(join(__dirname, '../../pages/admin/applications/index.vue'), 'utf8')

describe('BFF Taiga application import', () => {
  it('proxies GET unlinked projects with auth headers', () => {
    expect(unlinkedSrc).toContain('apiAuthHeaders(event)')
    expect(unlinkedSrc).toContain('/api/v1/integrations/taiga/projects/unlinked')
  })

  it('proxies GET link by application with auth headers', () => {
    expect(byApplicationSrc).toContain('apiAuthHeaders(event)')
    expect(byApplicationSrc).toContain('/api/v1/integrations/taiga/links/by-application/')
  })

  it('proxies POST bulk import with auth headers', () => {
    expect(importSrc).toContain('apiAuthHeaders(event)')
    expect(importSrc).toContain('/api/v1/integrations/taiga/applications/import')
    expect(importSrc).toMatch(/method:\s*['"]POST['"]/)
  })

  it('exposes Taiga import UI on admin applications page', () => {
    expect(appsPageSrc).toContain('applications.taiga_import_title')
    expect(appsPageSrc).toContain('taigaProjectId')
    expect(appsPageSrc).toContain('/api/integrations/taiga/projects/unlinked')
  })

  it('exposes Taiga link section on application edit modal', () => {
    expect(appsPageSrc).toContain('applications.taiga_link_title')
    expect(appsPageSrc).toContain('/api/integrations/taiga/links/by-application/')
    expect(appsPageSrc).toContain('existingTaigaLink')
  })
})
