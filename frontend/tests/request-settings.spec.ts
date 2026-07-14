import { describe, expect, it } from 'vitest'
import type { ChannelsEnabled } from '~/composables/useRequestSettings'

function activeChannelCount(ch: ChannelsEnabled) {
  return Number(ch.tma) + Number(ch.support) + Number(ch.maintenance)
}

describe('request channel settings', () => {
  it('counts active channels', () => {
    expect(activeChannelCount({ tma: true, support: false, maintenance: true })).toBe(2)
    expect(activeChannelCount({ tma: true, support: false, maintenance: false })).toBe(1)
  })
})
