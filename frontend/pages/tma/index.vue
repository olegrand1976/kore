<template>
  <div>
    <AppPageHeader :title="$t('tma.title')">
      <template #actions>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/tma/gantt')">
          {{ $t('tma.gantt') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="exportXml">
          {{ $t('tma.export') }}
        </AppButton>
        <AppButton variant="primary" size="sm" @click="showForm = !showForm">
          {{ $t('tma.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppKpiGrid compact>
      <AppKpiCard
        icon="list_alt"
        tone="gold"
        :loading="pending"
        :value="kpi.total"
        :label="$t('tma.kpi_total')"
      />
      <AppKpiCard
        icon="pending"
        tone="blue"
        :loading="pending"
        :value="kpi.open"
        :label="$t('tma.kpi_open')"
      />
      <AppKpiCard
        icon="check_circle"
        tone="success"
        :loading="pending"
        :value="kpi.resolved"
        :label="$t('tma.kpi_resolved')"
      />
      <AppKpiCard
        v-if="canValidateTma"
        icon="hourglass_empty"
        tone="warn"
        :loading="pending"
        :value="kpi.awaiting"
        :label="$t('tma.kpi_awaiting')"
      />
    </AppKpiGrid>

    <AppCard v-if="showForm" padding="lg" class="mb">
      <DemandForm @submit="onCreate" />
    </AppCard>

    <AppListToolbar
      v-if="!pending"
      :filters="listFilters"
      :filter-values="filterValues"
      :sort-keys="sortKeys"
      :sort-key="sortKey"
      :sort-dir="sortDir"
      :view="view"
      kanban-enabled
      :has-active-filters="hasActiveFilters"
      @update:filter="setFilter"
      @update:sort-key="setSort($event)"
      @update:sort-dir="setSortDir"
      @update:view="setView"
      @reset="resetFilters"
    />

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('tma.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="!displayRows.length" padding="lg">
      <AppEmptyState
        icon="inbox"
        :title="hasActiveFilters ? $t('common.list.no_results') : $t('tma.empty')"
      />
    </AppCard>

    <AppCard v-else-if="view === 'table'" padding="none">
      <AppTable :columns="columns" :rows="displayRows" row-key="id">
        <template #cell-status="{ value }">
          <AppBadge variant="neutral">{{ tmaStatusLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="navigateTo(`/tma/${row.id}`)">
            {{ $t('tma.open') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>

    <AppCard v-else padding="lg">
      <AppKanbanBoard
        :columns="kanbanColumns"
        :items="displayRows"
        :column-key="(row) => String((row as TmaRow).status)"
        :item-key="(row) => String((row as TmaRow).id)"
        :empty-label="$t('common.list.no_results')"
      >
        <template #card="{ item }">
          <div class="tma-kanban-card">
            <p class="tma-kanban-card__title">{{ (item as TmaRow).title }}</p>
            <AppBadge variant="neutral">{{ tmaStatusLabel(String((item as TmaRow).status)) }}</AppBadge>
            <AppButton variant="ghost" size="sm" @click="navigateTo(`/tma/${(item as TmaRow).id}`)">
              {{ $t('tma.open') }}
            </AppButton>
          </div>
        </template>
      </AppKanbanBoard>
    </AppCard>
  </div>
</template>

<script setup lang="ts">
import type { KanbanColumn } from '~/components/ui/AppKanbanBoard.vue'
import { countTmaByStatus, countTmaOpen } from '~/composables/useKpiMetrics'
import { applyTextSearch, useListControls } from '~/composables/useListControls'

definePageMeta({ layout: 'default', middleware: 'cra-gate' })

const TMA_STATUSES = [
  'en_attente_creation',
  'ouverte',
  'affectee',
  'en_cours',
  'rework',
  'resolue'
] as const

type TmaRow = {
  id: string
  title: string
  status: string
  createdAt: string
}

const { t } = useI18n()
const { list, create, exportXml, pickId, pickSubject, pickStatus, pickCreatedAt } = useTma()
const { canValidateTma } = usePermissions()

const showForm = ref(false)
const creating = ref(false)

const { data, pending, refresh } = await useAsyncData('tma-demands', () => list())

const listItems = computed((): TmaRow[] =>
  (data.value ?? []).map((d) => ({
    id: pickId(d),
    title: pickSubject(d),
    status: pickStatus(d),
    createdAt: pickCreatedAt(d)
  }))
)

const tmaStatusLabel = (status: string) => {
  const key = `dashboard.charts.status.tma.${status}` as const
  const translated = t(key)
  return translated === key ? status : translated
}

const listFilters = computed(() => ({
  status: {
    type: 'select' as const,
    label: t('tma.col_status'),
    options: TMA_STATUSES.map((status) => ({
      value: status,
      label: tmaStatusLabel(status)
    })),
    match: (row: TmaRow, value: string) => row.status === value
  },
  q: {
    type: 'search' as const,
    label: t('common.list.search'),
    placeholder: t('tma.col_title'),
    match: (row: TmaRow, query: string) => applyTextSearch(query, row.title)
  }
}))

const sortKeys = computed(() => [
  { key: 'createdAt', label: t('tma.sort_created'), type: 'date' as const, accessor: (row: TmaRow) => row.createdAt },
  { key: 'title', label: t('tma.col_title'), type: 'string' as const, accessor: (row: TmaRow) => row.title },
  { key: 'status', label: t('tma.col_status'), type: 'string' as const, accessor: (row: TmaRow) => row.status }
])

const {
  filterValues,
  sortKey,
  sortDir,
  view,
  sortedItems,
  hasActiveFilters,
  setFilter,
  setSort,
  setSortDir,
  setView,
  resetFilters
} = useListControls(listItems, {
  storageKey: 'tma-demands',
  defaultSort: { key: 'createdAt', dir: 'desc' },
  kanbanEnabled: true,
  filters: listFilters,
  sortKeys
})

const kpi = computed(() => {
  const items = data.value ?? []
  return {
    total: items.length,
    open: countTmaOpen(items),
    resolved: countTmaByStatus(items, 'resolue'),
    awaiting: countTmaByStatus(items, 'en_attente_creation')
  }
})

const columns = computed(() => [
  { key: 'title', label: t('tma.col_title') },
  { key: 'status', label: t('tma.col_status') },
  { key: 'actions', label: '' }
])

const displayRows = computed(() => sortedItems.value)

const kanbanColumns = computed((): KanbanColumn[] =>
  TMA_STATUSES.map((status) => ({
    id: status,
    label: tmaStatusLabel(status),
    tone: status === 'resolue' ? 'success' : status === 'en_attente_creation' ? 'warn' : 'blue'
  }))
)

const onCreate = async (payload: { applicationId: string; subject: string; requiresChefGate: boolean }) => {
  creating.value = true
  try {
    await create(payload)
    showForm.value = false
    await refresh()
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }

.muted {
  margin: 0;
  color: var(--kore-text-muted);
}

.tma-kanban-card {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
}

.tma-kanban-card__title {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text);
  word-break: break-word;
}
</style>
