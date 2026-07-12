import { defineStore } from 'pinia'

export type CraLine = {
  sourceType: string
  sourceId: string
  day: string
  duration: number
  comment?: string
}

export type CraWeek = {
  weekNumber: number
  submittedAt?: string | null
  lines: CraLine[]
}

export type CraTimesheet = {
  id: string
  month: string
  status: string
  commercialInfo?: { client?: string; mission?: string }
  weeks?: CraWeek[]
}

function normalizeWeek(raw: Record<string, unknown>): CraWeek {
  const weekNumber = Number(raw.weekNumber ?? raw.WeekNumber ?? 0)
  const linesRaw = (raw.lines ?? raw.Lines ?? []) as Record<string, unknown>[]
  const lines: CraLine[] = linesRaw.map((line) => {
    const source = (line.source ?? line.Source ?? {}) as Record<string, unknown>
    const duration = (line.duration ?? line.Duration ?? {}) as Record<string, unknown>
    const day = String(line.day ?? line.Day ?? '')
    return {
      sourceType: String(source.type ?? source.Type ?? 'manual'),
      sourceId: String(source.id ?? source.ID ?? 'default'),
      day: day.slice(0, 10),
      duration: Number(duration.minutes ?? duration.Minutes ?? line.duration ?? 0),
      comment: String(line.comment ?? line.Comment ?? '')
    }
  })
  return {
    weekNumber,
    submittedAt: (raw.submittedAt ?? raw.SubmittedAt ?? null) as string | null,
    lines
  }
}

function normalizeTimesheet(raw: Record<string, unknown>): CraTimesheet {
  const weeksRaw = (raw.weeks ?? raw.Weeks ?? []) as Record<string, unknown>[]
  return {
    id: String(raw.id ?? raw.ID ?? ''),
    month: String(raw.month ?? raw.Month ?? ''),
    status: String(raw.status ?? raw.Status ?? ''),
    commercialInfo: (raw.commercialInfo ?? raw.CommercialInfo) as CraTimesheet['commercialInfo'],
    weeks: weeksRaw.map(normalizeWeek)
  }
}

export const useCraStore = defineStore('cra', {
  state: () => ({
    timesheet: null as CraTimesheet | null,
    loading: false,
    saving: false,
    error: null as string | null
  }),
  getters: {
    canEdit: (state) => state.timesheet?.status !== 'Définitif',
    selectedWeeks: (state) => state.timesheet?.weeks ?? []
  },
  actions: {
    setTimesheet(raw: Record<string, unknown>) {
      this.timesheet = normalizeTimesheet(raw)
    },
    async load(id: string) {
      const { apiFetch } = useApiFetch()
      this.loading = true
      this.error = null
      try {
        const res = await apiFetch<{ data?: Record<string, unknown> }>(`/api/cra/timesheets/${id}`)
        const data = res.data ?? (res as unknown as Record<string, unknown>)
        this.setTimesheet(data)
      } catch {
        this.error = 'load_failed'
        this.timesheet = null
      } finally {
        this.loading = false
      }
    },
    async saveWeek(weekNumber: number, lines: CraLine[]) {
      if (!this.timesheet) return
      const { apiFetch } = useApiFetch()
      this.saving = true
      try {
        const res = await apiFetch<{ data?: Record<string, unknown> }>(
          `/api/cra/timesheets/${this.timesheet.id}/weeks/${weekNumber}`,
          { method: 'PUT', body: { lines } }
        )
        const data = res.data ?? (res as unknown as Record<string, unknown>)
        this.setTimesheet(data)
      } finally {
        this.saving = false
      }
    },
    async submitWeek(weekNumber: number) {
      if (!this.timesheet) return
      const { apiFetch } = useApiFetch()
      await apiFetch(`/api/cra/timesheets/${this.timesheet.id}/weeks/${weekNumber}/submit`, { method: 'POST' })
      await this.load(this.timesheet.id)
    },
    async validateFinal() {
      if (!this.timesheet) return
      const { apiFetch } = useApiFetch()
      await apiFetch(`/api/cra/timesheets/${this.timesheet.id}/validate`, { method: 'POST' })
      await this.load(this.timesheet.id)
    }
  }
})
