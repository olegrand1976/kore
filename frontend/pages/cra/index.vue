<template>
  <div>
    <AppPageHeader :title="$t('cra.title')">
      <template #actions>
        <AppButton variant="primary" size="sm" :disabled="creating" @click="openCurrentMonth">
          <AppIcon name="add" /> {{ $t('cra.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppKpiGrid compact>
      <AppKpiCard
        icon="list_alt"
        tone="gold"
        :loading="pending"
        :value="kpi.total"
        :label="$t('cra.kpi_total')"
      />
      <AppKpiCard
        icon="edit_note"
        tone="warn"
        :loading="pending"
        :value="kpi.drafts"
        :label="$t('cra.kpi_drafts')"
      />
      <AppKpiCard
        v-if="canValidateCra"
        icon="pending_actions"
        tone="blue"
        :loading="pending"
        :value="kpi.submitted"
        :label="$t('cra.kpi_submitted')"
      />
      <AppKpiCard
        v-else
        icon="today"
        tone="blue"
        :loading="pending"
        :value="kpi.currentStatusLabel"
        :label="$t('cra.kpi_current')"
        :hint="kpi.currentMonthLabel"
      />
      <AppKpiCard
        icon="check_circle"
        tone="success"
        :loading="pending"
        :value="kpi.finalized"
        :label="$t('cra.kpi_finalized')"
      />
    </AppKpiGrid>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('cra.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="rows.length" padding="none" class="cra-table-wrap">
      <AppTable :columns="columns" :rows="rows" row-key="id">
        <template #cell-month="{ value }">
          <span class="cra-month">{{ formatMonth(String(value)) }}</span>
        </template>
        <template #cell-userLogin="{ value }">
          <span class="cra-user">{{ formatUserLogin(String(value)) }}</span>
        </template>
        <template #cell-context="{ row }">
          <span class="cra-context" :class="{ 'cra-context--empty': !row.context }">
            {{ row.context || $t('cra.context_empty') }}
          </span>
        </template>
        <template #cell-hours="{ value }">
          <span class="cra-hours">{{ $t('cra.hours_value', { n: formatHours(Number(value)) }) }}</span>
        </template>
        <template #cell-status="{ value }">
          <AppBadge :variant="statusVariant(String(value))">{{ statusLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-updatedAt="{ value }">
          <span class="cra-updated">{{ formatUpdated(String(value)) }}</span>
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/cra/${row.id}`)">{{ $t('cra.open') }}</AppButton>
        </template>
      </AppTable>
    </AppCard>

    <AppCard v-else padding="lg">
      <AppEmptyState icon="schedule" :title="$t('cra.empty')" :description="$t('cra.empty_desc')">
        <AppButton variant="primary" size="sm" :disabled="creating" @click="openCurrentMonth">{{ $t('cra.new') }}</AppButton>
      </AppEmptyState>
    </AppCard>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
  </div>
</template>

<script setup lang="ts">
import { countCraByStatus } from '~/composables/useKpiMetrics'

definePageMeta({ layout: 'default' })

type CraSummary = {
  id: string
  userId?: string
  userLogin?: string
  month: string
  status: string
  commercialInfo?: { client?: string; mission?: string }
  totalMinutes?: number
  weeksSubmitted?: number
  updatedAt?: string
}

const { t, locale } = useI18n()
const { statusLabel, statusVariant, currentMonthKey } = useCraStatus()
const { canValidateCra } = usePermissions()

const creating = ref(false)
const errorMsg = ref('')

const { data, pending, refresh } = await useFetch('/api/cra/timesheets/recent')

const rawItems = computed((): CraSummary[] => {
  const payload = (data.value as { data?: unknown[] })?.data ?? data.value
  if (!Array.isArray(payload)) return []
  return payload.map((ts: Record<string, unknown>) => ({
    id: String(ts.id ?? ''),
    userId: ts.userId ? String(ts.userId) : undefined,
    userLogin: ts.userLogin ? String(ts.userLogin) : undefined,
    month: String(ts.month ?? ''),
    status: String(ts.status ?? ''),
    commercialInfo: (ts.commercialInfo as CraSummary['commercialInfo']) ?? undefined,
    totalMinutes: Number(ts.totalMinutes ?? 0),
    weeksSubmitted: Number(ts.weeksSubmitted ?? 0),
    updatedAt: ts.updatedAt ? String(ts.updatedAt) : undefined
  }))
})

const columns = computed(() => {
  const cols = [{ key: 'month', label: t('cra.col_period') }]
  if (canValidateCra.value) {
    cols.push({ key: 'userLogin', label: t('cra.col_user') })
  }
  cols.push(
    { key: 'context', label: t('cra.col_context') },
    { key: 'hours', label: t('cra.col_hours') },
    { key: 'status', label: t('cra.col_status') },
    { key: 'updatedAt', label: t('cra.col_updated') },
    { key: 'actions', label: '' }
  )
  return cols
})

const rows = computed(() =>
  rawItems.value.map((ts) => ({
    id: ts.id,
    month: ts.month,
    userLogin: ts.userLogin ?? '—',
    context: formatContext(ts.commercialInfo),
    hours: ts.totalMinutes ?? 0,
    status: ts.status,
    updatedAt: ts.updatedAt ?? '',
    actions: ''
  }))
)

const kpi = computed(() => {
  const items = rawItems.value
  const key = currentMonthKey()
  const current = items.find((ts) => ts.month === key)
  const [y, m] = key.split('-').map(Number)
  const currentMonthLabel = new Date(y, m - 1, 1).toLocaleDateString(
    locale.value === 'en' ? 'en-US' : 'fr-FR',
    { month: 'long', year: 'numeric' }
  )
  return {
    total: items.length,
    drafts: countCraByStatus(items, 'Brouillon'),
    submitted: countCraByStatus(items, 'ValidéSemaine'),
    finalized: countCraByStatus(items, 'Définitif'),
    currentStatusLabel: current ? statusLabel(current.status) : '—',
    currentMonthLabel
  }
})

const formatMonth = (raw: string) => {
  const [y, m] = raw.split('-').map(Number)
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long',
    year: 'numeric'
  })
}

const formatUserLogin = (login: string) => {
  if (!login || login === '—') return '—'
  const idx = login.indexOf('_')
  return idx >= 0 ? login.slice(idx + 1) : login
}

const formatContext = (info?: { client?: string; mission?: string }) => {
  if (!info) return ''
  const parts = [info.client, info.mission].filter(Boolean)
  return parts.join(' / ')
}

const formatHours = (minutes: number) => {
  if (!minutes) return '0'
  const hours = minutes / 60
  return Number.isInteger(hours) ? String(hours) : hours.toFixed(1)
}

const formatUpdated = (raw: string) => {
  if (!raw) return '—'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

const openCurrentMonth = async () => {
  creating.value = true
  errorMsg.value = ''
  try {
    const res = await $fetch<any>(`/api/cra/timesheets?month=${currentMonthKey()}`)
    const ts = res?.data ?? res
    if (ts?.id) {
      await navigateTo(`/cra/${ts.id}`)
      return
    }
    await refresh()
  } catch {
    errorMsg.value = t('cra.download_error')
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.cra-table-wrap { overflow: hidden; }

.cra-month {
  font-weight: 600;
  color: var(--kore-text);
}

.cra-user {
  font-weight: 500;
  color: var(--kore-text);
}

.cra-context {
  color: var(--kore-text);
  max-width: 14rem;
  display: inline-block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.cra-context--empty {
  color: var(--kore-text-muted);
  font-style: italic;
}

.cra-hours {
  font-variant-numeric: tabular-nums;
  color: var(--kore-text);
}

.cra-updated {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-caption);
  white-space: nowrap;
}

.muted { color: var(--kore-text-muted); }

.flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
}

.flash--error { color: var(--kore-error); }

@media (max-width: 768px) {
  .cra-context {
    max-width: 8rem;
  }
}
</style>
