import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const tmaDetailSrc = readFileSync(join(__dirname, '../pages/tma/[id].vue'), 'utf8')

describe('tma detail reopen reason', () => {
  it('displays reopen reason when status is rework', () => {
    expect(tmaDetailSrc).toContain('tma.reopen_reason')
    expect(tmaDetailSrc).toContain("status.value !== 'rework'")
    expect(tmaDetailSrc).toContain('pickReopenReason')
  })
})
