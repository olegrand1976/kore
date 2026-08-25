<template>
  <div>
    <AppPageHeader :title="$t('tma.title')" :subtitle="$t('tma.subtitle')">
      <template #actions>
        <AppButton v-if="guideRef?.dismissed" variant="ghost" size="sm" type="button" @click="guideRef?.showAgain()">
          {{ $t('guides.show') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="navigateTo('/tma/gantt')">
          {{ $t('tma.gantt') }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="exportXml">
          {{ $t('tma.export') }}
        </AppButton>
        <AppButton variant="primary" size="sm" @click="toggleForm">
          {{ $t('tma.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="tma" />

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

    <p v-if="flashOk" class="flash flash--ok" role="status">{{ flashOk }}</p>
    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>

    <AppCard v-if="showForm" padding="lg" class="mb">
      <ServiceRequestForm show-chef-gate show-agile-fields show-assignee :busy="creating" @submit="onCreate" />
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
        <template #cell-priority="{ value }">
          <AppBadge variant="neutral">{{ priorityLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-status="{ value }">
          <AppBadge variant="neutral">{{ tmaStatusLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-actions="{ row }">
          <div class="tma-actions">
            <AppButton
              variant="ghost"
              size="icon"
              :aria-label="$t('tma.open')"
              @click="navigateTo(`/tma/${row.id}`)"
            >
              <AppIcon name="visibility" />
            </AppButton>
            <AppButton
              v-if="canWriteTma"
              variant="ghost"
              size="icon"
              :aria-label="$t('tma.delete')"
              :disabled="deleting"
              @click="askDelete(row as TmaRow)"
            >
              <AppIcon name="delete" />
            </AppButton>
          </div>
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
            <p v-if="(item as TmaRow).applicationId" class="tma-kanban-card__meta">
              {{ (item as TmaRow).application }}
            </p>
            <p v-if="(item as TmaRow).assigneeId" class="tma-kanban-card__meta">
              {{ (item as TmaRow).assignee }}
            </p>
            <p v-if="(item as TmaRow).takenOverById" class="tma-kanban-card__meta">
              {{ (item as TmaRow).takenOverBy }}
            </p>
            <div class="tma-kanban-card__badges">
              <AppBadge variant="neutral">{{ priorityLabel(String((item as TmaRow).priority)) }}</AppBadge>
              <AppBadge variant="neutral">{{ tmaStatusLabel(String((item as TmaRow).status)) }}</AppBadge>
            </div>
            <div class="tma-actions">
              <AppButton
                variant="ghost"
                size="icon"
                :aria-label="$t('tma.open')"
                @click="navigateTo(`/tma/${(item as TmaRow).id}`)"
              >
                <AppIcon name="visibility" />
              </AppButton>
              <AppButton
                v-if="canWriteTma"
                variant="ghost"
                size="icon"
                :aria-label="$t('tma.delete')"
                :disabled="deleting"
                @click="askDelete(item as TmaRow)"
              >
                <AppIcon name="delete" />
              </AppButton>
            </div>
          </div>
        </template>
      </AppKanbanBoard>
    </AppCard>

    <AppModal
      v-model:open="deleteOpen"
      width="md"
      title-id="tma-delete-title"
      :aria-label="$t('tma.delete')"
    >
      <form class="tma-delete" @submit.prevent="confirmDelete">
        <h2 id="tma-delete-title" class="tma-delete__title">{{ $t('tma.delete') }}</h2>
        <p>{{ $t('tma.delete_confirm', { title: pendingDelete?.title || '' }) }}</p>
        <p v-if="deleteError" class="tma-delete__error" role="alert">{{ deleteError }}</p>
        <div class="tma-delete__actions">
          <AppButton variant="ghost" size="sm" type="button" :disabled="deleting" @click="deleteOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="danger" size="sm" type="submit" :disabled="deleting">
            {{ $t('tma.delete') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import type { KanbanColumn } from '~/components/ui/AppKanbanBoard.vue'
import type { ServiceRequestPayload } from '~/components/requests/ServiceRequestForm.vue'
import { countTmaByStatus, countTmaOpen } from '~/composables/useKpiMetrics'
import { applyTextSearch, useListControls } from '~/composables/useListControls'
import { REQUEST_RESOURCE, useRequestAttachments } from '~/composables/useRequestAttachments'

definePageMeta({ layout: 'default' })

const TMA_STATUSES = [
  'en_attente_creation',
  'ouverte',
  'affectee',
  'en_cours',
  'rework',
  'resolue'
] as const

const TMA_PRIORITIES = ['low', 'normal', 'high', 'urgent'] as const

type TmaRow = {
  id: string
  title: string
  status: string
  applicationId: string
  application: string
  assigneeId: string
  assignee: string
  takenOverById: string
  takenOverBy: string
  priority: string
  createdAt: string
}

const { t } = useI18n()
const route = useRoute()
const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)
const { extractFetchError } = useApiError()
const {
  list,
  create,
  remove,
  exportXml,
  pickId,
  pickSubject,
  pickStatus,
  pickPriority,
  pickApplicationId,
  pickAssigneeId,
  pickTakenOverById,
  pickCreatedAt
} = useTma()
const { uploadAll } = useRequestAttachments()
const { can, canValidateTma } = usePermissions()
const { list: listApps, pickAppLabel, appById } = useApplications()
const { list: listUsers, pickUserId, pickUserLogin } = useUsers()

const canWriteTma = computed(() => can('tma', 'E'))

const showForm = ref(false)
const creating = ref(false)
const errorMsg = ref('')
const flashOk = ref('')
const deleteOpen = ref(false)
const deleting = ref(false)
const deleteError = ref('')
const pendingDelete = ref<TmaRow | null>(null)

const toggleForm = () => {
  showForm.value = !showForm.value
  if (!showForm.value) {
    errorMsg.value = ''
  }
}

if (route.query.create === '1') {
  showForm.value = true
}

const { data, pending, refresh } = await useAsyncData('tma-demands', () => list())
// Label enrichment only — Collaborateur may lack org L for /users.
const { data: appsData } = await useAsyncData('tma-apps', () => listApps().catch(() => []))
const { data: usersData } = await useAsyncData('tma-users', () => listUsers().catch(() => []))

const appMap = computed(() => appById(appsData.value ?? []))

const userLoginById = computed(() => {
  const map = new Map<string, string>()
  for (const user of usersData.value ?? []) {
    const id = pickUserId(user)
    if (!id) continue
    map.set(id, pickUserLogin(user) || id)
  }
  return map
})

const listItems = computed((): TmaRow[] =>
  (data.value ?? []).map((d) => {
    const applicationId = pickApplicationId(d)
    const assigneeId = pickAssigneeId(d)
    const takenOverById = pickTakenOverById(d)
    return {
      id: pickId(d),
      title: pickSubject(d),
      status: pickStatus(d),
      applicationId,
      application: pickAppLabel(appMap.value.get(applicationId)) || applicationId || t('common.none'),
      assigneeId,
      assignee: assigneeId
        ? (userLoginById.value.get(assigneeId) || assigneeId)
      : t('common.none'),
      takenOverById,
      takenOverBy: takenOverById
        ? (userLoginById.value.get(takenOverById) || takenOverById)
        : t('common.none'),
      priority: pickPriority(d),
      createdAt: pickCreatedAt(d),
    }
  })
)

const tmaStatusLabel = (status: string) => {
  const key = `dashboard.charts.status.tma.${status}` as const
  const translated = t(key)
  return translated === key ? status : translated
}

const priorityLabel = (priority: string) =>
  t(`requests.priority_${priority}` as const, priority)

const appFilterOptions = computed(() => {
  const seen = new Map<string, string>()
  for (const row of listItems.value) {
    if (!row.applicationId || seen.has(row.applicationId)) continue
    seen.set(row.applicationId, row.application)
  }
  return [...seen.entries()]
    .map(([value, label]) => ({ value, label }))
    .sort((a, b) => a.label.localeCompare(b.label))
})

const listFilters = computed(() => ({
  applicationId: {
    type: 'select' as const,
    label: t('requests.col_application'),
    options: appFilterOptions.value,
    match: (row: TmaRow, value: string) => row.applicationId === value
  },
  status: {
    type: 'select' as const,
    label: t('tma.col_status'),
    options: TMA_STATUSES.map((status) => ({
      value: status,
      label: tmaStatusLabel(status)
    })),
    match: (row: TmaRow, value: string) => row.status === value
  },
  priority: {
    type: 'select' as const,
    label: t('tma.col_priority'),
    options: TMA_PRIORITIES.map((priority) => ({
      value: priority,
      label: priorityLabel(priority)
    })),
    match: (row: TmaRow, value: string) => row.priority === value
  },
  q: {
    type: 'search' as const,
    label: t('common.list.search'),
    placeholder: t('tma.search_placeholder'),
    match: (row: TmaRow, query: string) =>
      applyTextSearch(query, row.title, row.application, row.assignee, row.takenOverBy)
  }
}))

const sortKeys = computed(() => [
  { key: 'createdAt', label: t('tma.sort_created'), type: 'date' as const, accessor: (row: TmaRow) => row.createdAt },
  { key: 'title', label: t('tma.col_title'), type: 'string' as const, accessor: (row: TmaRow) => row.title },
  { key: 'application', label: t('requests.col_application'), type: 'string' as const, accessor: (row: TmaRow) => row.application },
  { key: 'assignee', label: t('requests.col_assignee'), type: 'string' as const, accessor: (row: TmaRow) => row.assignee },
  { key: 'takenOverBy', label: t('requests.col_taken_over_by'), type: 'string' as const, accessor: (row: TmaRow) => row.takenOverBy },
  { key: 'priority', label: t('tma.col_priority'), type: 'string' as const, accessor: (row: TmaRow) => row.priority },
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
  { key: 'application', label: t('requests.col_application') },
  { key: 'assignee', label: t('requests.col_assignee') },
  { key: 'takenOverBy', label: t('requests.col_taken_over_by') },
  { key: 'priority', label: t('tma.col_priority') },
  { key: 'status', label: t('tma.col_status') },
  { key: 'actions', label: t('tma.col_actions') }
])

const displayRows = computed(() => sortedItems.value)

const kanbanColumns = computed((): KanbanColumn[] =>
  TMA_STATUSES.map((status) => ({
    id: status,
    label: tmaStatusLabel(status),
    tone: status === 'resolue' ? 'success' : status === 'en_attente_creation' ? 'warn' : 'blue'
  }))
)

const askDelete = (row: TmaRow) => {
  pendingDelete.value = row
  flashOk.value = ''
  errorMsg.value = ''
  deleteError.value = ''
  deleteOpen.value = true
}

const confirmDelete = async () => {
  const row = pendingDelete.value
  if (!row?.id) return
  deleting.value = true
  deleteError.value = ''
  flashOk.value = ''
  try {
    await remove(row.id)
    deleteOpen.value = false
    pendingDelete.value = null
    flashOk.value = t('tma.delete_ok')
    await refresh()
  } catch (e) {
    deleteError.value = extractFetchError(e, t('tma.delete_error'))
  } finally {
    deleting.value = false
  }
}

const onCreate = async (payload: ServiceRequestPayload) => {
  creating.value = true
  errorMsg.value = ''
  try {
    const created = await create({
      applicationId: payload.applicationId,
      assigneeId: payload.assigneeId,
      subject: payload.subject,
      description: payload.description,
      priority: payload.priority,
      dueAt: payload.dueAt,
      requiresChefGate: payload.requiresChefGate,
      epicId: payload.epicId,
      storyPoints: payload.storyPoints
    })
    const id = pickId(created)
    if (id && payload.files.length) {
      await uploadAll(REQUEST_RESOURCE.tma, id, payload.files)
    }
    showForm.value = false
    await refresh()
  } catch (e) {
    errorMsg.value = extractFetchError(e, t('tma.error_create'))
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

.flash {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-small);
}

.flash--error { color: var(--kore-error); }
.flash--ok { color: var(--kore-success); }

.tma-actions {
  display: inline-flex;
  align-items: center;
  gap: var(--kore-space-xs);
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

.tma-kanban-card__meta {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  word-break: break-word;
}

.tma-kanban-card__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.tma-delete {
  display: grid;
  gap: var(--kore-space-md);
}

.tma-delete__title {
  margin: 0;
  font-size: var(--kore-text-h3);
  color: var(--kore-text);
}

.tma-delete p {
  margin: 0;
  color: var(--kore-text);
}

.tma-delete__error {
  color: var(--kore-error);
  font-size: var(--kore-text-small);
}

.tma-delete__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}

@media (max-width: 768px) {
  .tma-delete__actions {
    flex-direction: column;
  }

  .tma-delete__actions :deep(.app-btn) {
    width: 100%;
  }
}
</style>
