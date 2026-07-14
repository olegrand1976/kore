export type ReportingPeriod = {
  start: string
  end: string
  window?: string
  Start?: string
  End?: string
}

export type CraGanttItem = {
  id: string
  label: string
  start: Date
  end: Date
  progress: number
}

export type PlanningSlot = {
  date: string
  label: string
  hours: number
}

export type PlanningRow = {
  userId: string
  userName: string
  slots: PlanningSlot[]
}

function parseDate(raw: string | undefined): Date | null {
  if (!raw) return null
  const d = new Date(raw)
  return Number.isNaN(d.getTime()) ? null : d
}

function monthPeriod(date = new Date()) {
  const start = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), 1))
  const end = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth() + 1, 0))
  const fmt = (d: Date) => d.toISOString().slice(0, 10)
  return { start: fmt(start), end: fmt(end) }
}

function rollingWindow60(date = new Date()) {
  const start = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate()))
  const end = new Date(start)
  end.setUTCDate(end.getUTCDate() + 59)
  const fmt = (d: Date) => d.toISOString().slice(0, 10)
  return { window: '60', start: fmt(start), end: fmt(end) }
}

export function useReporting() {
  const fetchGantt = async (params?: { start?: string; end?: string; window?: string }) => {
    const period = params?.window === '60' || (!params?.start && !params?.end)
      ? rollingWindow60()
      : params?.start && params?.end
        ? params
        : monthPeriod()
    const res = await $fetch<{
      data?: {
        period?: ReportingPeriod
        Period?: ReportingPeriod
        items?: Array<Record<string, unknown>>
        Items?: Array<Record<string, unknown>>
      }
    }>('/api/gantt', { query: period })
    const data = res.data ?? (res as unknown as Record<string, unknown>)
    const itemsRaw = (data.items ?? data.Items ?? []) as Array<Record<string, unknown>>
    const items: CraGanttItem[] = itemsRaw
      .map((item) => {
        const start = parseDate(String(item.startDate ?? item.StartDate ?? ''))
        const end = parseDate(String(item.endDate ?? item.EndDate ?? ''))
        if (!start || !end) return null
        return {
          id: String(item.id ?? item.ID ?? ''),
          label: String(item.label ?? item.Label ?? ''),
          start,
          end,
          progress: Number(item.progress ?? item.Progress ?? 0)
        }
      })
      .filter((item): item is CraGanttItem => item != null && item.id !== '')
    return items
  }

  const fetchPlanning = async (params?: { start?: string; end?: string; window?: string }) => {
    const period = params?.window === '60' || (!params?.start && !params?.end)
      ? rollingWindow60()
      : params?.start && params?.end
        ? params
        : monthPeriod()
    const res = await $fetch<{
      data?: {
        rows?: Array<Record<string, unknown>>
        Rows?: Array<Record<string, unknown>>
      }
    }>('/api/planning', { query: period })
    const data = res.data ?? (res as unknown as Record<string, unknown>)
    const rowsRaw = (data.rows ?? data.Rows ?? []) as Array<Record<string, unknown>>
    return rowsRaw.map((row): PlanningRow => {
      const slotsRaw = (row.slots ?? row.Slots ?? []) as Array<Record<string, unknown>>
      return {
        userId: String(row.userId ?? row.UserID ?? ''),
        userName: String(row.userName ?? row.UserName ?? ''),
        slots: slotsRaw.map((slot) => ({
          date: String(slot.date ?? slot.Date ?? '').slice(0, 10),
          label: String(slot.label ?? slot.Label ?? ''),
          hours: Number(slot.hours ?? slot.Hours ?? 0)
        }))
      }
    })
  }

  const fetchBillingStats = async (params?: { start?: string; end?: string; window?: string }) => {
    const period = params?.window === '60' ? rollingWindow60() : monthPeriod()
    const res = await $fetch<{
      data?: {
        totalAmount?: number
        TotalAmount?: number
        invoiceCount?: number
        InvoiceCount?: number
        billableHours?: number
        BillableHours?: number
        currency?: string
        Currency?: string
      }
    }>('/api/billing-stats', { query: period })
    const data = res.data ?? (res as unknown as Record<string, unknown>)
    return {
      totalAmount: Number(data.totalAmount ?? data.TotalAmount ?? 0),
      invoiceCount: Number(data.invoiceCount ?? data.InvoiceCount ?? 0),
      billableHours: Number(data.billableHours ?? data.BillableHours ?? 0),
      currency: String(data.currency ?? data.Currency ?? 'EUR')
    }
  }

  return { fetchGantt, fetchPlanning, fetchBillingStats, monthPeriod, rollingWindow60 }
}
