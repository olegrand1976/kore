<template>
  <div>
    <AppPageHeader :title="$t('budget.title')" />
    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('budget.empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'default' })

const { t } = useI18n()
const { data, pending } = await useFetch('/api/budget/budgets')

const items = computed(() => (data.value as any)?.data ?? [])

const columns = computed(() => [
  { key: 'type', label: t('budget.col_type') },
  { key: 'planned', label: t('budget.col_planned') },
  { key: 'consumed', label: t('budget.col_consumed') },
  { key: 'currency', label: t('budget.col_currency') }
])

const rows = computed(() =>
  items.value.map((b: any) => ({
    type: b.type ?? b.Type,
    planned: b.planned?.days ?? b.Planned?.Days ?? '-',
    consumed: b.consumed?.days ?? b.Consumed?.Days ?? '-',
    currency: b.currency ?? b.Currency ?? 'EUR',
    _link: `/budget/${b.id ?? b.ID}`
  }))
)
</script>
