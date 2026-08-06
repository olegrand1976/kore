<template>
  <div>
    <AppPageHeader :title="$t('budget.title')" :subtitle="$t('budget.subtitle')">
      <template #actions>
        <AppButton v-if="guideRef?.dismissed" variant="ghost" size="sm" type="button" @click="guideRef?.showAgain()">
          {{ $t('guides.show') }}
        </AppButton>
        <AppButton
          v-if="canManageOrg"
          variant="ghost"
          size="sm"
          type="button"
          @click="navigateTo('/admin/applications')"
        >
          {{ $t('budget.manage_applications') }}
        </AppButton>
        <AppButton
          v-if="canWriteBudget"
          variant="primary"
          size="sm"
          type="button"
          @click="openCreateModal"
        >
          <AppIcon name="add" /> {{ $t('budget.create') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="budget" />

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
        :hint="kpi.consumptionPct > 0 ? $t('budget.kpi_consumption_pct', { n: kpi.consumptionPct }) : undefined"
      />
      <AppKpiCard
        icon="warning"
        :tone="kpi.overrun > 0 ? 'warn' : 'success'"
        :loading="pending"
        :value="kpi.overrun"
        :label="$t('budget.kpi_overrun')"
        :hint="kpi.overrun > 0 ? $t('budget.kpi_overrun_hint', { n: kpi.overrun }) : undefined"
      />
    </AppKpiGrid>

    <AppCard v-if="loadError" padding="lg">
      <AppEmptyState icon="error" :title="loadError" />
    </AppCard>

    <template v-else>
      <AppListToolbar
        :filters="listFilters"
        :filter-values="filterValues"
        :sort-keys="sortKeys"
        :sort-key="sortKey"
        :sort-dir="sortDir"
        :has-active-filters="hasActiveFilters"
        @update:filter="setFilter"
        @update:sort-key="setSort($event)"
        @update:sort-dir="setSortDir"
        @reset="resetFilters"
      />

      <AppCard padding="lg">
        <AppTable
          :columns="columns"
          :rows="displayRows"
          :loading="pending"
          :empty-title="hasActiveFilters ? $t('common.list.no_results') : $t('budget.empty')"
          :empty-description="hasActiveFilters ? undefined : $t('budget.empty_desc')"
        >
        <template #empty>
          <div v-if="!hasActiveFilters" class="budget-empty-actions">
            <AppButton v-if="canWriteBudget" variant="primary" type="button" @click="openCreateModal">
              <AppIcon name="add" /> {{ $t('budget.create') }}
            </AppButton>
            <AppButton
              v-if="canManageOrg"
              variant="ghost"
              type="button"
              @click="navigateTo('/admin/applications')"
            >
              {{ $t('budget.manage_applications') }}
            </AppButton>
          </div>
        </template>
        <template #cell-application="{ row }">
          <button type="button" class="row-link" @click="navigateTo(`/budget/${row.id}`)">
            {{ row.application }}
          </button>
        </template>
        <template #cell-client="{ value }">
          <span :class="{ muted: !value }">{{ value || $t('budget.col_empty') }}</span>
        </template>
        <template #cell-type="{ row }">
          <AppBadge variant="gold">{{ row.typeLabel }}</AppBadge>
        </template>
        <template #cell-consumption="{ row }">
          <div class="consumption-cell">
            <div class="consumption-cell__track" role="progressbar" :aria-valuenow="row.consumptionPct" aria-valuemin="0" aria-valuemax="100">
              <div
                class="consumption-cell__fill"
                :class="`consumption-cell__fill--${row.status}`"
                :style="{ width: `${Math.min(100, row.consumptionPct)}%` }"
              />
            </div>
            <span class="consumption-cell__pct">{{ row.consumptionPct }} %</span>
            <AppBadge v-if="row.status === 'overrun'" variant="error">{{ $t('budget.status_overrun') }}</AppBadge>
          </div>
        </template>
        <template #cell-days="{ row }">
          {{ row.consumed }} / {{ row.planned }} {{ $t('budget.unit_days') }}
        </template>
        <template #cell-actions="{ row }">
          <div class="budget-row-actions">
            <AppButton variant="ghost" size="sm" @click="navigateTo(`/budget/${row.id}`)">
              {{ $t('budget.open') }}
            </AppButton>
            <AppButton
              v-if="canManageOrg && row.applicationId"
              variant="ghost"
              size="sm"
              type="button"
              @click="openEditApp(row)"
            >
              {{ $t('budget.edit_application') }}
            </AppButton>
          </div>
        </template>
      </AppTable>
      </AppCard>
    </template>

    <AppModal v-model:open="createOpen" width="sm" :aria-label="$t('budget.create_title')">
      <form class="budget-form" @submit.prevent="submitCreate">
        <h2 class="budget-form__title">{{ $t('budget.create_title') }}</h2>
        <AppApplicationSelect
          id="budget-create-app"
          v-model="createForm.applicationId"
          :label="$t('budget.form_application')"
          required
        />
        <div class="budget-form__field">
          <label for="budget-create-type">{{ $t('budget.form_type') }}</label>
          <select id="budget-create-type" v-model="createForm.type" required>
            <option value="defaut">{{ $t('budget.type_defaut') }}</option>
            <option value="specifique">{{ $t('budget.type_specifique') }}</option>
          </select>
          <p class="budget-form__hint">{{ $t('budget.type_defaut_help') }}</p>
        </div>
        <AppInput
          id="budget-create-days"
          v-model="createForm.plannedDays"
          type="number"
          step="0.5"
          min="0"
          :label="$t('budget.form_planned_days')"
          required
        />
        <AppInput
          id="budget-create-uo"
          v-model="createForm.plannedUO"
          type="number"
          step="0.5"
          min="0"
          :label="$t('budget.form_planned_uo')"
          required
        />
        <AppInput
          id="budget-create-amount"
          v-model="createForm.plannedAmountEur"
          type="number"
          step="0.01"
          min="0"
          :label="$t('budget.form_planned_amount_eur')"
          required
        />
        <p v-if="createError" class="budget-form__error" role="alert">{{ createError }}</p>
        <div class="budget-form__actions">
          <AppButton variant="ghost" type="button" @click="createOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :disabled="creating">
            {{ creating ? $t('budget.creating') : $t('budget.create_submit') }}
          </AppButton>
        </div>
      </form>
    </AppModal>

    <AppModal v-model:open="editAppOpen" width="sm" :aria-label="$t('org.tree.edit_application_title')">
      <form class="budget-form" @submit.prevent="submitEditApp">
        <h2 class="budget-form__title">{{ $t('org.tree.edit_application_title') }}</h2>
        <AppInput
          id="budget-edit-app-libelle"
          v-model="editAppForm.libelle"
          :label="$t('org.tree.field_libelle')"
          required
        />
        <p v-if="editAppError" class="budget-form__error" role="alert">{{ editAppError }}</p>
        <div class="budget-form__actions">
          <AppButton variant="ghost" type="button" @click="editAppOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :disabled="editAppSaving">
            {{ editAppSaving ? $t('org.tree.saving') : $t('org.tree.save') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { budgetMetrics, consumptionPct } from '~/composables/useKpiMetrics'
import { useListControls } from '~/composables/useListControls'

definePageMeta({ layout: 'default' })

const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)

type BudgetRow = {
  id: string
  applicationId: string
  application: string
  client: string
  typeLabel: string
  type: string
  planned: number
  consumed: number
  consumptionPct: number
  status: string
}

const { t } = useI18n()
const { list, create, pickId, tripleValue } = useBudget()
const {
  list: listApplications,
  update: updateApplication,
  appById,
  pickAppLabel,
  pickAppClient
} = useApplications()
const { budgetTypeLabel, budgetStatus, consumptionPercent } = useBudgetDisplay()
const { extractFetchError } = useApiError()
const { can } = usePermissions()
const { isAdmin } = useAuth()

const canWriteBudget = computed(() => can('budget', 'E'))
const canManageOrg = computed(() => isAdmin.value && can('org', 'E'))

const loadError = ref('')
const createOpen = ref(false)
const creating = ref(false)
const createError = ref('')
const createForm = reactive({
  applicationId: '',
  type: 'defaut' as 'defaut' | 'specifique',
  plannedDays: '0',
  plannedUO: '0',
  plannedAmountEur: '0'
})

const editAppOpen = ref(false)
const editAppSaving = ref(false)
const editAppError = ref('')
const editAppForm = reactive({ id: '', libelle: '' })

const { data, pending, refresh } = await useAsyncData('budget-list', async () => {
  loadError.value = ''
  try {
    const [budgets, applications] = await Promise.all([list(), listApplications({ active: 'all' })])
    return { budgets, applications }
  } catch (err) {
    loadError.value = extractFetchError(err)
    return { budgets: [], applications: [] }
  }
})

const openCreateModal = (presetApplicationId = '') => {
  createForm.applicationId = presetApplicationId
  createForm.type = 'defaut'
  createForm.plannedDays = '0'
  createForm.plannedUO = '0'
  createForm.plannedAmountEur = '0'
  createError.value = ''
  createOpen.value = true
}

const submitCreate = async () => {
  creating.value = true
  createError.value = ''
  try {
    const plannedAmountEur = Number(createForm.plannedAmountEur) || 0
    await create({
      applicationId: createForm.applicationId,
      type: createForm.type,
      plannedDays: Number(createForm.plannedDays) || 0,
      plannedUO: Number(createForm.plannedUO) || 0,
      plannedAmount: Math.round(plannedAmountEur * 100),
      currency: 'EUR'
    })
    createOpen.value = false
    await refresh()
  } catch (err) {
    createError.value = extractFetchError(err, t('budget.create_error'))
  } finally {
    creating.value = false
  }
}

const openEditApp = (row: BudgetRow) => {
  editAppForm.id = row.applicationId
  editAppForm.libelle = row.application
  editAppError.value = ''
  editAppOpen.value = true
}

const submitEditApp = async () => {
  editAppSaving.value = true
  editAppError.value = ''
  try {
    await updateApplication(editAppForm.id, { libelle: editAppForm.libelle })
    editAppOpen.value = false
    await refresh()
  } catch (err) {
    editAppError.value = extractFetchError(err, t('org.tree.update_error'))
  } finally {
    editAppSaving.value = false
  }
}

const appMap = computed(() => appById(data.value?.applications ?? []))

const kpi = computed(() => {
  const m = budgetMetrics(data.value?.budgets ?? [])
  return {
    total: m.total,
    plannedDays: m.plannedDays,
    consumedDays: m.consumedDays,
    overrun: m.overrun,
    consumptionPct: consumptionPct(m.consumedDays, m.plannedDays, false)
  }
})

const listItems = computed((): BudgetRow[] =>
  (data.value?.budgets ?? []).map((b) => {
    const id = pickId(b)
    const appId = b.applicationId ?? b.ApplicationID ?? ''
    const app = appMap.value.get(appId)
    const planned = tripleValue(b.planned ?? b.Planned, 'days')
    const consumed = tripleValue(b.consumed ?? b.Consumed, 'days')
    const type = b.type ?? b.Type ?? ''
    const status = budgetStatus(consumed, planned)
    return {
      id,
      applicationId: appId,
      application: pickAppLabel(app) || id.slice(0, 8),
      client: pickAppClient(app) || '',
      typeLabel: budgetTypeLabel(type),
      type,
      planned,
      consumed,
      consumptionPct: consumptionPercent(consumed, planned),
      status
    }
  })
)

const budgetTypes = computed(() => {
  const types = new Set<string>()
  for (const row of listItems.value) {
    if (row.type) types.add(row.type)
  }
  return [...types]
})

const listFilters = computed(() => ({
  type: {
    type: 'select' as const,
    label: t('budget.col_type'),
    options: budgetTypes.value.map((type) => ({
      value: type,
      label: budgetTypeLabel(type)
    })),
    match: (row: BudgetRow, value: string) => row.type === value
  },
  consumption: {
    type: 'select' as const,
    label: t('budget.col_consumption'),
    options: [
      { value: 'ok', label: t('budget.status_ok') },
      { value: 'warn', label: t('budget.status_warn') },
      { value: 'overrun', label: t('budget.status_overrun') }
    ],
    match: (row: BudgetRow, value: string) => row.status === value
  }
}))

const sortKeys = computed(() => [
  {
    key: 'consumptionPct',
    label: t('budget.col_consumption'),
    type: 'number' as const,
    accessor: (row: BudgetRow) => row.consumptionPct
  },
  {
    key: 'application',
    label: t('budget.col_application'),
    type: 'string' as const,
    accessor: (row: BudgetRow) => row.application
  },
  {
    key: 'client',
    label: t('budget.col_client'),
    type: 'string' as const,
    accessor: (row: BudgetRow) => row.client
  }
])

const {
  filterValues,
  sortKey,
  sortDir,
  sortedItems,
  hasActiveFilters,
  setFilter,
  setSort,
  setSortDir,
  resetFilters
} = useListControls(listItems, {
  storageKey: 'budget-list',
  defaultSort: { key: 'consumptionPct', dir: 'desc' },
  filters: listFilters,
  sortKeys
})

const displayRows = computed(() => sortedItems.value)

const columns = computed(() => [
  { key: 'application', label: t('budget.col_application') },
  { key: 'client', label: t('budget.col_client') },
  { key: 'type', label: t('budget.col_type') },
  { key: 'consumption', label: t('budget.col_consumption') },
  { key: 'days', label: t('budget.col_days') },
  { key: 'actions', label: '' }
])

const route = useRoute()
onMounted(() => {
  if (route.query.create === '1') {
    const appId = typeof route.query.applicationId === 'string' ? route.query.applicationId : ''
    openCreateModal(appId)
  }
})
</script>

<style scoped>
.row-link {
  background: none;
  border: none;
  padding: 0;
  font: inherit;
  color: var(--kore-accent);
  cursor: pointer;
  text-align: left;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.muted {
  color: var(--kore-text-muted);
}

.consumption-cell {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
  min-width: 8rem;
}

.consumption-cell__track {
  flex: 1 1 4rem;
  height: 0.4rem;
  background: var(--kore-bg-subtle);
  border-radius: var(--kore-radius-full);
  overflow: hidden;
}

.consumption-cell__fill {
  height: 100%;
  border-radius: var(--kore-radius-full);
}

.consumption-cell__fill--ok {
  background: var(--kore-accent);
}

.consumption-cell__fill--warn {
  background: var(--kore-brand-gold);
}

.consumption-cell__fill--overrun {
  background: var(--kore-danger);
}

.consumption-cell__pct {
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  white-space: nowrap;
}

.budget-row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.budget-empty-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: center;
  margin-top: var(--kore-space-lg);
}

.budget-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}

.budget-form__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.budget-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.budget-form__field label {
  font-size: var(--kore-text-small);
  font-weight: 500;
}

.budget-form__field select {
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}

.budget-form__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.budget-form__error {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-error);
}

.budget-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}

@media (max-width: 768px) {
  .budget-form__actions {
    flex-direction: column-reverse;
  }

  .budget-form__actions :deep(.app-button),
  .budget-empty-actions :deep(.app-button),
  .budget-row-actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
