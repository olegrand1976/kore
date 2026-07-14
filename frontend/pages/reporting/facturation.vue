<template>
  <div>
    <AppPageHeader :title="$t('reporting.facturation_title')" :subtitle="$t('reporting.facturation_subtitle')" />

    <AppCard padding="lg">
      <p v-if="pending" class="muted">{{ $t('cra.loading') }}</p>
      <p v-else-if="error" class="flash flash--error" role="alert">{{ error }}</p>
      <dl v-else class="stats-grid">
        <div>
          <dt>{{ $t('reporting.billable_hours') }}</dt>
          <dd>{{ stats?.billableHours?.toFixed(1) ?? '0' }} h</dd>
        </div>
        <div>
          <dt>{{ $t('reporting.invoice_count') }}</dt>
          <dd>{{ stats?.invoiceCount ?? 0 }}</dd>
        </div>
        <div>
          <dt>{{ $t('reporting.total_amount') }}</dt>
          <dd>{{ formatAmount(stats?.totalAmount ?? 0, stats?.currency ?? 'EUR') }}</dd>
        </div>
      </dl>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { fetchBillingStats } = useReporting()
const { t } = useI18n()

const { data: stats, pending, error: fetchError } = await useAsyncData('billing-stats', () => fetchBillingStats({ window: '60' }))

const error = computed(() => (fetchError.value ? t('reporting.facturation_error') : ''))

const formatAmount = (cents: number, currency: string) => {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(cents / 100)
}
</script>

<style scoped>
.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: var(--kore-space-lg);
  margin: 0;
}

.stats-grid dt {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.stats-grid dd {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-h2);
  font-weight: 600;
}
</style>
