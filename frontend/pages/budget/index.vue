<template>
  <div>
    <AppPageHeader :title="$t('budget.title')" />

    <AppKpiGrid compact>
      <AppKpiCard
        icon="folder"
        tone="gold"
        :loading="pending"
        :value="kpi.total"
        :label="$t('budget.kpi_total')"
      />
      <AppKpiCard
        icon="event_available"
        tone="blue"
        :loading="pending"
        :value="kpi.plannedDays"
        :label="$t('budget.kpi_planned_days')"
      />
      <AppKpiCard
        icon="trending_up"
        tone="success"
        :loading="pending"
        :value="kpi.consumedDays"
        :label="$t('budget.kpi_consumed_days')"
      />
      <AppKpiCard
        icon="warning"
        :tone="kpi.overrun > 0 ? 'warn' : 'success'"
        :loading="pending"
        :value="kpi.overrun"
        :label="$t('budget.kpi_overrun')"
        :hint="kpi.consumptionPct > 0 ? $t('budget.kpi_consumption_pct', { n: kpi.consumptionPct }) : undefined"
      />
    </AppKpiGrid>

    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('budget.empty')"
      >
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/budget/${row.id}`)">
            {{ $t('budget.open') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import { budgetMetrics, consumptionPct } from '~/composables/useKpiMetrics'

definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { list, pickId, tripleValue } = useBudget()

const { data, pending } = await useAsyncData('budget-list', () => list())

const kpi = computed(() => {
  const m = budgetMetrics(data.value ?? [])
  return {
    total: m.total,
    plannedDays: m.plannedDays,
    consumedDays: m.consumedDays,
    overrun: m.overrun,
    consumptionPct: consumptionPct(m.consumedDays, m.plannedDays)
  }
})

const columns = computed(() => [
  { key: 'type', label: t('budget.col_type') },
  { key: 'planned', label: t('budget.col_planned') },
  { key: 'consumed', label: t('budget.col_consumed') },
  { key: 'currency', label: t('budget.col_currency') },
  { key: 'actions', label: '' }
])

const rows = computed(() =>
  (data.value ?? []).map((b) => ({
    id: pickId(b),
    type: b.type ?? b.Type,
    planned: tripleValue(b.planned ?? b.Planned, 'days'),
    consumed: tripleValue(b.consumed ?? b.Consumed, 'days'),
    currency: b.currency ?? b.Currency ?? 'EUR'
  }))
)
</script>
