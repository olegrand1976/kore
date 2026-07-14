<template>
  <div>
    <AppPageHeader :title="$t('reporting.tma_title')" :subtitle="$t('reporting.tma_subtitle')" />

    <AppCard padding="lg">
      <p v-if="pending" class="muted">{{ $t('cra.loading') }}</p>
      <p v-else-if="error" class="flash flash--error" role="alert">{{ $t('reporting.tma_error') }}</p>
      <AppTable
        v-else
        :columns="columns"
        :rows="rows"
        :empty-title="$t('reporting.tma_empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t } = useI18n()

const columns = computed(() => [
  { key: 'metric', label: t('reporting.tma_col_metric') },
  { key: 'value', label: t('reporting.tma_col_value') }
])

const { data, pending, error } = await useAsyncData('report-tma-summary', () =>
  $fetch<{ data?: { rows?: Array<Record<string, unknown>>; Rows?: Array<Record<string, unknown>> } }>(
    '/api/reports/run',
    { method: 'POST', body: { reportCode: 'tma_summary', params: {} } }
  )
)

const rows = computed(() => {
  const raw = data.value?.data?.rows ?? data.value?.data?.Rows ?? []
  return raw.map((row, index) => ({
    id: String(index),
    metric: String(row.metric ?? row.Metric ?? '—'),
    value: String(row.value ?? row.Value ?? '—')
  }))
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
