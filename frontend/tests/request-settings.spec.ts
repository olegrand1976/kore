import { describe, expect, it } from 'vitest'
import type { ChannelsEnabled, TenantRequestSettings } from '~/composables/useRequestSettings'

function activeChannelCount(ch: ChannelsEnabled) {
  return Number(ch.tma) + Number(ch.support) + Number(ch.maintenance)
}

function isInvoicingEnabled(settings: Pick<TenantRequestSettings, 'invoicingEnabled'> | null, loaded: boolean) {
  if (!loaded) return false
  return settings?.invoicingEnabled ?? false
}

describe('request channel settings', () => {
  it('counts active channels', () => {
    expect(activeChannelCount({ tma: true, support: false, maintenance: true })).toBe(2)
    expect(activeChannelCount({ tma: true, support: false, maintenance: false })).toBe(1)
  })

  it('gates invoicing menu on org flag', () => {
    expect(isInvoicingEnabled(null, false)).toBe(false)
    expect(isInvoicingEnabled({ invoicingEnabled: false }, true)).toBe(false)
    expect(isInvoicingEnabled({ invoicingEnabled: true }, true)).toBe(true)
  })
})

function toInvoiceLinePayload(line: {
  description: string
  quantity: string
  unitPriceEur: string
  taxRate: string
}) {
  return {
    description: line.description.trim(),
    quantity: Number(line.quantity),
    unitPrice: Math.round(Number(line.unitPriceEur) * 100),
    taxRate: Number(line.taxRate)
  }
}

describe('manual invoice creation payload', () => {
  it('converts euro unit price to cents', () => {
    expect(
      toInvoiceLinePayload({
        description: ' Presta ',
        quantity: '2.5',
        unitPriceEur: '120.50',
        taxRate: '20'
      })
    ).toEqual({
      description: 'Presta',
      quantity: 2.5,
      unitPrice: 12050,
      taxRate: 20
    })
  })
})
