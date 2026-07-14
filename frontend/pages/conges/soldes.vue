<template>
  <div>
    <h3 class="conges-subtitle">{{ $t('conges.balances_title') }}</h3>
    <AppListToolbar
      :filter-values="{}"
      :sort-keys="sortKeys"
      :sort-key="sortKey"
      :sort-dir="sortDir"
      @update:sort-key="setSort($event)"
      @update:sort-dir="setSortDir"
    />
    <AppCard padding="lg">
      <AppTable
        :columns="columns"
        :rows="displayRows"
        :loading="pending"
        :empty-title="$t('conges.balances_empty')"
      />
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import { useListControls } from '~/composables/useListControls'

const { t } = useI18n()
const { balances } = useLeave()
const { fetchMine } = useLeaveTypeConfigs()
const { typeLabel } = useLeaveLabels()

await fetchMine()

const { data, pending } = await useAsyncData('leave-balances-page', () => balances())

type BalanceRow = {
  type: string
  acquired: string
  taken: string
  remaining: string
  remainingNum: number | null
}

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

const listItems = computed((): BalanceRow[] =>
  (data.value ?? []).map((item) => {
    const remainingRaw = item.remaining ?? item.Remaining ?? item.balance ?? item.Balance
    const remainingNum = remainingRaw == null || Number.isNaN(Number(remainingRaw)) ? null : Number(remainingRaw)
    return {
      type: typeLabel(item.type ?? item.Type ?? ''),
      acquired: formatBalance(item.acquired ?? item.Acquired),
      taken: formatBalance(item.taken ?? item.Taken),
      remaining: formatBalance(remainingNum),
      remainingNum
    }
  })
)

const sortKeys = computed(() => [
  {
    key: 'remaining',
    label: t('conges.balances_amount'),
    type: 'number' as const,
    accessor: (row: BalanceRow) => row.remainingNum ?? -1
  },
  { key: 'type', label: t('conges.col_type'), type: 'string' as const, accessor: (row: BalanceRow) => row.type }
])

const { sortKey, sortDir, sortedItems, setSort, setSortDir } = useListControls(listItems, {
  storageKey: 'leave-balances',
  defaultSort: { key: 'remaining', dir: 'asc' },
  sortKeys
})

const displayRows = computed(() => sortedItems.value)
</script>

<style scoped>
.conges-subtitle {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
  font-weight: 600;
  color: var(--kore-text);
}
</style>
