<template>
  <div>
    <h3 class="conges-subtitle">{{ $t('conges.balances_title') }}</h3>
    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="rows"
        :loading="pending"
        :empty-title="$t('conges.balances_empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n()
const { balances } = useLeave()
const { fetchMine } = useLeaveTypeConfigs()
const { typeLabel } = useLeaveLabels()

await fetchMine()

const { data, pending } = await useAsyncData('leave-balances-page', () => balances())

const columns = computed(() => [
  { key: 'type', label: t('conges.col_type') },
  { key: 'acquired', label: t('conges.balances_acquired') },
  { key: 'taken', label: t('conges.balances_taken') },
  { key: 'remaining', label: t('conges.balances_amount') }
])

const formatBalance = (value: number | null | undefined) => {
  if (value == null || Number.isNaN(value)) return '—'
  return Number.isInteger(value) ? String(value) : value.toFixed(1)
}

const rows = computed(() =>
  (data.value ?? []).map((item) => ({
    type: typeLabel(item.type ?? item.Type ?? ''),
    acquired: formatBalance(item.acquired ?? item.Acquired),
    taken: formatBalance(item.taken ?? item.Taken),
    remaining: formatBalance(item.remaining ?? item.Remaining ?? item.balance ?? item.Balance)
  }))
)
</script>

<style scoped>
.conges-subtitle {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
  font-weight: 600;
  color: var(--kore-text);
}
</style>
