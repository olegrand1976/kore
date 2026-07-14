<template>
  <div>
    <AppPageHeader :title="$t('reporting.dashboard_title', { code })" />

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('cra.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="error" padding="lg">
      <p class="flash flash--error" role="alert">{{ $t('reporting.dashboard_error') }}</p>
    </AppCard>

    <DashboardGrid v-else :items="items" />
  </div>
</template>

<script setup lang="ts">
import type { DashboardGridItem } from '~/components/reporting/DashboardGrid.vue'

definePageMeta({ layout: 'default' })

const route = useRoute()
const { t } = useI18n()
const code = computed(() => String(route.params.code ?? 'cra'))

const { data, pending, error } = await useAsyncData(
  () => `dashboard-${code.value}`,
  () => $fetch<{ data?: Record<string, unknown> }>(`/api/dashboards/${code.value}`)
)

const items = computed((): DashboardGridItem[] => {
  const payload = data.value?.data?.payload ?? data.value?.data?.Payload ?? data.value?.data ?? {}
  const record = payload as Record<string, unknown>
  const currency = String(record.currency ?? record.Currency ?? 'EUR')
  const totalAmount = Number(record.totalAmount ?? record.TotalAmount ?? 0)
  const formatter = new Intl.NumberFormat(undefined, { style: 'currency', currency })
  return [
    {
      key: 'billable',
      icon: 'schedule',
      tone: 'blue',
      value: Number(record.billableHours ?? record.BillableHours ?? 0).toFixed(1),
      label: t('reporting.billable_hours'),
      to: '/cra/planning'
    },
    {
      key: 'invoices',
      icon: 'receipt_long',
      tone: 'gold',
      value: Number(record.invoiceCount ?? record.InvoiceCount ?? 0),
      label: t('reporting.invoice_count'),
      to: '/reporting/facturation'
    },
    {
      key: 'amount',
      icon: 'payments',
      tone: 'success',
      value: formatter.format(totalAmount / 100),
      label: t('reporting.total_amount'),
      to: '/reporting/facturation'
    }
  ]
})
</script>

<style scoped>
.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.flash--error {
  color: var(--kore-error);
}
</style>
