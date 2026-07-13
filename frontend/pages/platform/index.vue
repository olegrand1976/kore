<template>
  <div class="platform-page">
    <AppPageHeader :title="$t('platform.title')" :subtitle="$t('platform.subtitle')" />

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('platform.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('platform.forbidden')" />
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <AppEmptyState icon="error" :title="$t('platform.error')" />
    </AppCard>

    <template v-else-if="overview">
      <AppKpiGrid>
        <AppKpiCard
          icon="corporate_fare"
          tone="gold"
          :value="overview.summary.totalTenants"
          :label="$t('platform.kpi.total_tenants')"
        />
        <AppKpiCard
          icon="trending_up"
          tone="success"
          :value="overview.summary.activeTenants30d"
          :label="$t('platform.kpi.active_tenants')"
          :hint="$t('platform.kpi.active_hint')"
        />
        <AppKpiCard
          icon="group"
          tone="blue"
          :value="overview.summary.totalActiveUsers"
          :label="$t('platform.kpi.total_users')"
          :hint="seatsHint"
        />
        <AppKpiCard
          icon="payments"
          tone="warn"
          :value="trialCount"
          :label="$t('platform.kpi.trial_tenants')"
        />
      </AppKpiGrid>

      <AppCard padding="lg" class="platform-page__table">
        <h3 class="platform-page__section-title">{{ $t('platform.tenants_title') }}</h3>
        <AppTable
          :columns="columns"
          :rows="rows"
          row-key="id"
          :empty-title="$t('platform.empty')"
        >
          <template #cell-societeName="{ value }">
            <span class="platform-page__tenant-name">{{ value }}</span>
          </template>
          <template #cell-subscriptionStatus="{ value }">
            <AppBadge :variant="statusVariant(value)">
              {{ statusLabel(value) }}
            </AppBadge>
          </template>
          <template #cell-activeUsers="{ row }">
            <span :class="{ 'platform-page__warn': seatWarn(row) }">
              {{ row.activeUsers }} / {{ row.seatLimit || '∞' }}
            </span>
          </template>
          <template #cell-activeLast30d="{ value }">
            <AppBadge :variant="value ? 'success' : 'default'">
              {{ value ? $t('platform.active_yes') : $t('platform.active_no') }}
            </AppBadge>
          </template>
          <template #cell-lastActivityAt="{ value }">
            {{ formatDate(value) }}
          </template>
        </AppTable>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: ['platform']
})

const { t, locale } = useI18n()
const { overview, pending, error, forbidden, fetchOverview } = usePlatformOverview()

onMounted(() => {
  fetchOverview()
})

const seatsHint = computed(() => {
  const limit = overview.value?.summary.totalSeatLimit ?? 0
  if (!limit) return undefined
  return t('platform.kpi.seats_hint', { n: limit })
})

const trialCount = computed(() => overview.value?.summary.tenantsByStatus?.trial ?? 0)

const columns = computed(() => [
  { key: 'societeName', label: t('platform.col.tenant') },
  { key: 'subscriptionStatus', label: t('platform.col.status'), nowrap: true },
  { key: 'activeUsers', label: t('platform.col.seats'), nowrap: true },
  { key: 'modulesEnabled', label: t('platform.col.modules'), nowrap: true },
  { key: 'craCount', label: t('platform.col.cra'), nowrap: true },
  { key: 'tmaOpen', label: t('platform.col.tma_open'), nowrap: true },
  { key: 'budgetCount', label: t('platform.col.budget'), nowrap: true },
  { key: 'aiRequests30d', label: t('platform.col.ai'), nowrap: true },
  { key: 'activeLast30d', label: t('platform.col.active'), nowrap: true },
  { key: 'lastActivityAt', label: t('platform.col.last_activity'), nowrap: true }
])

const rows = computed(() =>
  (overview.value?.tenants ?? []).map((tenant) => ({
    ...tenant,
    societeName: tenant.societeName || tenant.name
  }))
)

function statusLabel(status: unknown): string {
  const key = String(status ?? 'none')
  if (key === 'none') return t('platform.status.none')
  const known = ['trial', 'active', 'past_due', 'suspended', 'canceled'] as const
  if ((known as readonly string[]).includes(key)) {
    return t(`dashboard.modules_panel.status.${key}`)
  }
  return key
}

function statusVariant(status: unknown): 'success' | 'warning' | 'default' {
  const key = String(status ?? 'none')
  switch (key) {
    case 'active':
      return 'success'
    case 'past_due':
    case 'suspended':
      return 'warning'
    default:
      return 'default'
  }
}

function seatWarn(row: Record<string, unknown>): boolean {
  const limit = Number(row.seatLimit ?? 0)
  const users = Number(row.activeUsers ?? 0)
  return limit > 0 && users / limit >= 0.9
}

function formatDate(value: unknown): string {
  if (!value || typeof value !== 'string') return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleDateString(locale.value === 'fr' ? 'fr-FR' : 'en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}
</script>

<style scoped>
.platform-page__section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-lg);
  font-weight: 600;
}

.platform-page__table {
  margin-top: var(--kore-space-md);
}

.platform-page__tenant-name {
  font-weight: 500;
}

.platform-page__warn {
  color: var(--kore-warn);
  font-weight: 600;
}

.muted {
  color: var(--kore-text-muted);
}
</style>
