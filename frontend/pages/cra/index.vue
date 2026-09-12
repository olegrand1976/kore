<template>
  <div>
    <AppPageHeader :title="$t('cra.title')">
      <template #actions>
        <AppButton v-if="guideRef?.dismissed" variant="ghost" size="sm" type="button" @click="guideRef?.showAgain()">
          {{ $t('guides.show') }}
        </AppButton>
        <AppButton v-if="canReadReporting" variant="ghost" size="sm" @click="navigateTo('/cra/planning')">
          {{ $t('cra.planning_link') }}
        </AppButton>
        <AppButton v-if="canReadReporting" variant="ghost" size="sm" @click="navigateTo('/cra/gantt')">
          {{ $t('cra.gantt_link') }}
        </AppButton>
        <AppButton variant="secondary" size="sm" :disabled="creating" @click="openPeriodModal">
          <AppIcon name="history" /> {{ $t('cra.new_other_period') }}
        </AppButton>
        <AppButton variant="primary" size="sm" :disabled="creating" @click="openCurrentMonth">
          <AppIcon name="add" /> {{ $t('cra.new') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="cra" />

    <p v-if="successMsg" class="flash" role="status">{{ successMsg }}</p>
    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <p v-if="canValidateCra && orgUsersLoadFailed" class="flash flash--error" role="alert">
      {{ $t('cra.filter_users_load_error') }}
    </p>

    <AppKpiGrid compact>
      <AppKpiCard
        icon="list_alt"
        tone="gold"
        :loading="pending"
        :value="kpi.total"
        :label="$t('cra.kpi_total')"
      />
      <AppKpiCard
        icon="edit_note"
        tone="warn"
        :loading="pending"
        :value="kpi.drafts"
        :label="$t('cra.kpi_drafts')"
      />
      <AppKpiCard
        v-if="canValidateCra"
        icon="pending_actions"
        tone="blue"
        :loading="pending"
        :value="kpi.submitted"
        :label="$t('cra.kpi_submitted')"
      />
      <AppKpiCard
        v-else
        icon="today"
        tone="blue"
        :loading="pending"
        :value="kpi.currentStatusLabel"
        :label="$t('cra.kpi_current')"
        :hint="kpi.currentMonthLabel"
      />
      <AppKpiCard
        icon="check_circle"
        tone="success"
        :loading="pending"
        :value="kpi.finalized"
        :label="$t('cra.kpi_finalized')"
      />
    </AppKpiGrid>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('cra.loading') }}</p>
    </AppCard>

    <template v-else>
      <AppListToolbar
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

      <AppCard v-if="!displayRows.length" padding="lg">
        <AppEmptyState
          icon="schedule"
          :title="hasActiveFilters ? $t('common.list.no_results') : $t('cra.empty')"
          :description="hasActiveFilters ? undefined : $t('cra.empty_desc')"
        >
          <div class="cra-empty-actions">
            <AppButton v-if="!hasActiveFilters" variant="secondary" size="sm" :disabled="creating" @click="openPeriodModal">
              {{ $t('cra.new_other_period') }}
            </AppButton>
            <AppButton v-if="!hasActiveFilters" variant="primary" size="sm" :disabled="creating" @click="openCurrentMonth">
              {{ $t('cra.new') }}
            </AppButton>
          </div>
        </AppEmptyState>
      </AppCard>

      <AppCard v-else-if="view === 'table'" padding="none" class="cra-table-wrap">
        <AppTable :columns="columns" :rows="displayRows" row-key="id">
        <template #cell-month="{ value }">
          <span class="cra-month">{{ formatMonth(String(value)) }}</span>
        </template>
        <template #cell-user="{ row }">
          <NuxtLink
            v-if="row.userId"
            :to="`/collaborateurs/${row.userId}`"
            class="cra-link"
          >
            {{ row.userDisplay }}
          </NuxtLink>
          <span v-else class="cra-user">{{ row.userDisplay }}</span>
        </template>
        <template #cell-client="{ row }">
          <NuxtLink
            v-if="row.clientId"
            :to="`/clients/${row.clientId}`"
            class="cra-link cra-link--truncate"
          >
            {{ row.client || $t('cra.context_empty') }}
          </NuxtLink>
          <span
            v-else
            class="cra-context"
            :class="{ 'cra-context--empty': !row.client }"
          >
            {{ row.client || $t('cra.context_empty') }}
          </span>
        </template>
        <template #cell-mission="{ row }">
          <NuxtLink
            v-if="row.missionId"
            :to="`/missions/${row.missionId}`"
            class="cra-link cra-link--truncate"
          >
            {{ row.mission || $t('cra.context_empty') }}
          </NuxtLink>
          <span
            v-else
            class="cra-context"
            :class="{ 'cra-context--empty': !row.mission }"
          >
            {{ row.mission || $t('cra.context_empty') }}
          </span>
        </template>
        <template #cell-hours="{ value }">
          <span class="cra-hours">{{ $t('cra.hours_value', { n: minutesToHoursLabel(Number(value)) }) }}</span>
        </template>
        <template #cell-status="{ value }">
          <AppBadge :variant="statusVariant(String(value))">{{ statusLabel(String(value)) }}</AppBadge>
        </template>
        <template #cell-updatedAt="{ value }">
          <span class="cra-updated">{{ formatUpdated(String(value)) }}</span>
        </template>
        <template #cell-actions="{ row }">
          <div class="cra-actions">
            <AppButton variant="success" size="sm" :to="`/cra/${row.id}`">{{ $t('cra.open') }}</AppButton>
            <CraDeleteTimesheetControl
              :timesheet-id="String(row.id)"
              :status="String(row.status)"
              :user-label="String(row.userDisplay)"
              :month="String(row.month)"
              @changed="onTimesheetAdminChange"
              @error="onTimesheetDeleteError"
            />
          </div>
        </template>
      </AppTable>
      </AppCard>

      <AppCard v-else padding="lg">
        <AppKanbanBoard
          :columns="kanbanColumns"
          :items="displayRows"
          :column-key="(row) => String((row as CraRow).status)"
          :item-key="(row) => String((row as CraRow).id)"
          :empty-label="$t('common.list.no_results')"
        >
          <template #card="{ item }">
            <div class="cra-kanban-card">
              <p class="cra-kanban-card__title">{{ formatMonth(String((item as CraRow).month)) }}</p>
              <p v-if="canValidateCra" class="cra-kanban-card__meta">
                <NuxtLink
                  v-if="(item as CraRow).userId"
                  :to="`/collaborateurs/${(item as CraRow).userId}`"
                  class="cra-link"
                >
                  {{ (item as CraRow).userDisplay }}
                </NuxtLink>
                <span v-else>{{ (item as CraRow).userDisplay }}</span>
              </p>
              <p class="cra-kanban-card__meta">
                <NuxtLink
                  v-if="(item as CraRow).clientId"
                  :to="`/clients/${(item as CraRow).clientId}`"
                  class="cra-link"
                >
                  {{ (item as CraRow).client || $t('cra.context_empty') }}
                </NuxtLink>
                <span v-else>{{ (item as CraRow).client || $t('cra.context_empty') }}</span>
                · {{ $t('cra.hours_value', { n: minutesToHoursLabel(Number((item as CraRow).hours)) }) }}
              </p>
              <p v-if="(item as CraRow).missionId || (item as CraRow).mission" class="cra-kanban-card__meta">
                <NuxtLink
                  v-if="(item as CraRow).missionId"
                  :to="`/missions/${(item as CraRow).missionId}`"
                  class="cra-link"
                >
                  {{ (item as CraRow).mission || $t('cra.context_empty') }}
                </NuxtLink>
                <span v-else>{{ (item as CraRow).mission }}</span>
              </p>
              <AppBadge :variant="statusVariant(String((item as CraRow).status))">
                {{ statusLabel(String((item as CraRow).status)) }}
              </AppBadge>
              <AppButton variant="success" size="sm" :to="`/cra/${(item as CraRow).id}`">
                {{ $t('cra.open') }}
              </AppButton>
              <CraDeleteTimesheetControl
                :timesheet-id="(item as CraRow).id"
                :status="(item as CraRow).status"
                :user-label="(item as CraRow).userDisplay"
                :month="(item as CraRow).month"
                @changed="onTimesheetAdminChange"
                @error="onTimesheetDeleteError"
              />
            </div>
          </template>
        </AppKanbanBoard>
      </AppCard>
    </template>

    <AppModal
      v-model:open="periodModalOpen"
      width="sm"
      :title-id="periodModalTitleId"
      :aria-label="$t('cra.new_period_title')"
    >
      <form class="cra-period-form" @submit.prevent="confirmPeriodModal">
        <h2 :id="periodModalTitleId" class="cra-period-form__title">{{ $t('cra.new_period_title') }}</h2>
        <p class="cra-period-form__hint">{{ $t('cra.new_period_hint') }}</p>
        <div class="cra-period-form__row">
          <div class="cra-period-form__field">
            <label for="cra-period-month">{{ $t('cra.filter_month') }}</label>
            <select id="cra-period-month" v-model="periodForm.month" required>
              <option v-for="opt in periodMonthOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
          <div class="cra-period-form__field">
            <label for="cra-period-year">{{ $t('cra.filter_year') }}</label>
            <select id="cra-period-year" v-model="periodForm.year" required>
              <option v-for="opt in periodYearOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
        </div>
        <p v-if="periodModalError" class="cra-period-form__error" role="alert">{{ periodModalError }}</p>
        <div class="cra-period-form__actions">
          <AppButton variant="ghost" type="button" :disabled="creating" @click="periodModalOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :disabled="creating || !periodMonthOptions.length">
            {{ creating ? $t('cra.opening') : $t('cra.new_period_submit') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import type { KanbanColumn } from '~/components/ui/AppKanbanBoard.vue'
import { countCraByStatus } from '~/composables/useKpiMetrics'
import { useCraError } from '~/composables/useCraError'
import { currentMonthKey, useCraStatus } from '~/composables/useCraStatus'
import { applyTextSearch, useListControls, type FilterDef } from '~/composables/useListControls'
import { formatUserDisplayName } from '~/composables/useUserDisplay'
import { minutesToHoursLabel } from '~/composables/useWeekCalendar'
import {
  useUsers,
  type OrgUserSummary
} from '~/composables/useUsers'
import {
  buildCreatePeriodMonthOptions,
  buildCreatePeriodYearOptions,
  clampCreatePeriodSelection,
  monthKeyFromParts,
  previousMonthKey
} from '~/utils/craCreatePeriod'
import {
  buildMonthFilterOptions,
  buildYearFilterOptions,
  matchPeriodMonthYear,
  migrateLegacyMonthFilter
} from '~/utils/craPeriodFilter'
import { buildCraUserFilterOptions } from '~/utils/craUserFilter'

definePageMeta({ layout: 'default' })

const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)

const CRA_STATUSES = ['Brouillon', 'ValidéSemaine', 'Définitif'] as const

type CraRow = {
  id: string
  month: string
  userId: string
  userDisplay: string
  client: string
  mission: string
  clientId: string
  missionId: string
  hours: number
  status: string
  updatedAt: string
  actions: string
}

type CraSummary = {
  id: string
  userId?: string
  userLogin?: string
  userPrenom?: string
  userNom?: string
  month: string
  status: string
  commercialInfo?: { client?: string; mission?: string; missionId?: string; clientId?: string }
  clientId?: string
  missionId?: string
  totalMinutes?: number
  weeksSubmitted?: number
  updatedAt?: string
}

const { t, locale } = useI18n()
const { statusLabel, statusVariant } = useCraStatus()
const { mapCraError } = useCraError()
const { canValidateCra, canReadReporting } = usePermissions()
const { apiFetch } = useApiFetch()
const { user, fetchSession } = useAuth()

await fetchSession()

const creating = ref(false)
const errorMsg = ref('')
const successMsg = ref('')
const periodModalOpen = ref(false)
const periodModalError = ref('')
const periodModalTitleId = 'cra-period-modal-title'
const prevDefault = previousMonthKey()
const periodForm = reactive({
  year: prevDefault.slice(0, 4),
  month: prevDefault.slice(5, 7)
})
const periodYearOptions = computed(() => buildCreatePeriodYearOptions())
const periodMonthOptions = computed(() =>
  buildCreatePeriodMonthOptions(periodForm.year, locale.value)
)

watch(
  () => periodForm.year,
  (year) => {
    const opts = buildCreatePeriodMonthOptions(year, locale.value)
    if (opts.some((o) => o.value === periodForm.month)) return
    if (opts.length) {
      periodForm.month = opts[opts.length - 1]!.value
      return
    }
    const clamped = clampCreatePeriodSelection(year, periodForm.month)
    periodForm.year = clamped.year
    periodForm.month = clamped.month
  }
)

const currentKey = currentMonthKey()
const defaultPeriodYear = currentKey.slice(0, 4)
const defaultPeriodMonth = currentKey.slice(5, 7)
const sessionUserId = computed(() => user.value?.userId ?? '')

/** Query for server-side recent list (period + optional user). Seed userId before first fetch. */
const recentQuery = reactive({
  year: defaultPeriodYear,
  month: defaultPeriodMonth,
  userId: canValidateCra.value ? (user.value?.userId ?? '') : '',
  limit: '500'
})

const { data, pending, refresh } = await useFetch('/api/cra/timesheets/recent', {
  query: recentQuery,
  watch: [recentQuery]
})

/** Full collaborator directory for the user filter (validators only). */
const {
  data: orgUsersRaw,
  error: orgUsersError
} = await useFetch<{ data?: OrgUserSummary[] }>('/api/org/users', {
  immediate: canValidateCra.value
})

const {
  pickUserId,
  pickUserLogin,
  pickUserPrenom,
  pickUserNom,
  pickUserActive,
  pickUserCraRequis
} = useUsers()

const orgUsersLoadFailed = computed(() => Boolean(orgUsersError.value))

const rawItems = computed((): CraSummary[] => {
  const payload = (data.value as { data?: unknown[] })?.data ?? data.value
  if (!Array.isArray(payload)) return []
  return payload.map((ts: Record<string, unknown>) => {
    const commercialInfo = (ts.commercialInfo as CraSummary['commercialInfo']) ?? undefined
    return {
    id: String(ts.id ?? ''),
    userId: ts.userId ? String(ts.userId) : undefined,
    userLogin: ts.userLogin ? String(ts.userLogin) : undefined,
    userPrenom: ts.userPrenom ? String(ts.userPrenom) : undefined,
    userNom: ts.userNom ? String(ts.userNom) : undefined,
    month: String(ts.month ?? ''),
    status: String(ts.status ?? ''),
    commercialInfo,
    clientId: ts.clientId
      ? String(ts.clientId)
      : commercialInfo?.clientId
        ? String(commercialInfo.clientId)
        : undefined,
    missionId: ts.missionId
      ? String(ts.missionId)
      : commercialInfo?.missionId
        ? String(commercialInfo.missionId)
        : undefined,
    totalMinutes: Number(ts.totalMinutes ?? 0),
    weeksSubmitted: Number(ts.weeksSubmitted ?? 0),
    updatedAt: ts.updatedAt ? String(ts.updatedAt) : undefined
  }})
})

const listItems = computed((): CraRow[] =>
  rawItems.value.map((ts) => ({
    id: ts.id,
    month: ts.month,
    userId: ts.userId ?? '',
    userDisplay: formatUserDisplayName(ts.userPrenom, ts.userNom, ts.userLogin),
    client: ts.commercialInfo?.client ?? '',
    mission: ts.commercialInfo?.mission ?? '',
    clientId: ts.clientId ?? '',
    missionId: ts.missionId ?? '',
    hours: ts.totalMinutes ?? 0,
    status: ts.status,
    updatedAt: ts.updatedAt ?? '',
    actions: ''
  }))
)

const userFilterOptions = computed(() => {
  const payload = (orgUsersRaw.value as { data?: OrgUserSummary[] } | null)?.data ?? orgUsersRaw.value
  const users: Array<{
    id: string
    prenom?: string
    nom?: string
    login?: string
    active?: boolean
    craRequis?: boolean
  }> = []
  if (Array.isArray(payload)) {
    for (const row of payload) {
      const id = pickUserId(row)
      if (!id) continue
      users.push({
        id,
        prenom: pickUserPrenom(row) || undefined,
        nom: pickUserNom(row) || undefined,
        login: pickUserLogin(row) || undefined,
        active: pickUserActive(row),
        craRequis: pickUserCraRequis(row)
      })
    }
  }
  const extras = listItems.value
    .filter((row) => row.userId)
    .map((row) => ({ id: row.userId, label: row.userDisplay || row.userId }))
  return buildCraUserFilterOptions(users, extras, sessionUserId.value, t('cra.filter_user_me'))
})

const listFilters = computed(() => {
  const filters: Record<string, FilterDef<CraRow>> = {
    status: {
      type: 'select' as const,
      label: t('cra.col_status'),
      options: CRA_STATUSES.map((status) => ({
        value: status,
        label: statusLabel(status)
      })),
      match: (row: CraRow, value: string) => row.status === value
    },
    periodMonth: {
      type: 'select' as const,
      label: t('cra.filter_month'),
      options: buildMonthFilterOptions(locale.value),
      defaultValue: defaultPeriodMonth,
      match: (row: CraRow, value: string) => matchPeriodMonthYear(row.month, value, '')
    },
    periodYear: {
      type: 'select' as const,
      label: t('cra.filter_year'),
      options: buildYearFilterOptions(listItems.value.map((row) => row.month)),
      defaultValue: defaultPeriodYear,
      match: (row: CraRow, value: string) => matchPeriodMonthYear(row.month, '', value)
    },
    q: {
      type: 'search' as const,
      label: t('common.list.search'),
      placeholder: t('cra.search_placeholder'),
      match: (row: CraRow, query: string) =>
        applyTextSearch(query, row.userDisplay, row.client, row.mission)
    }
  }
  if (canValidateCra.value) {
    filters.user = {
      type: 'select' as const,
      label: t('cra.col_user'),
      options: userFilterOptions.value,
      defaultValue: sessionUserId.value,
      match: (row: CraRow, value: string) => row.userId === value
    }
  }
  return filters
})

const sortKeys = computed(() => {
  const keys = [
    { key: 'month', label: t('cra.col_period'), type: 'date' as const, accessor: (row: CraRow) => row.month },
    { key: 'updatedAt', label: t('cra.col_updated'), type: 'date' as const, accessor: (row: CraRow) => row.updatedAt },
    { key: 'hours', label: t('cra.col_hours'), type: 'number' as const, accessor: (row: CraRow) => row.hours }
  ]
  if (canValidateCra.value) {
    keys.splice(1, 0, {
      key: 'user',
      label: t('cra.col_user'),
      type: 'string' as const,
      accessor: (row: CraRow) => row.userDisplay
    })
  }
  return keys
})

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
  storageKey: 'cra-recent-v2',
  defaultSort: { key: 'month', dir: 'desc' },
  kanbanEnabled: true,
  filters: listFilters,
  sortKeys
})

migrateLegacyMonthFilter(filterValues)

const ensurePeriodDefaults = () => {
  if (!filterValues.periodYear) filterValues.periodYear = defaultPeriodYear
  if (!filterValues.periodMonth) filterValues.periodMonth = defaultPeriodMonth
  if (canValidateCra.value && sessionUserId.value && !filterValues.user) {
    filterValues.user = sessionUserId.value
  }
}
ensurePeriodDefaults()

watch(
  () => [filterValues.periodYear, filterValues.periodMonth, filterValues.user] as const,
  ([year, month, userId]) => {
    // Empty year/month = "Tous" : do not coerce back to current period.
    recentQuery.year = year || ''
    recentQuery.month = month || ''
    recentQuery.userId = canValidateCra.value ? (userId || '') : ''
  },
  { immediate: true }
)

const displayRows = computed(() => sortedItems.value)

const kanbanColumns = computed((): KanbanColumn[] =>
  CRA_STATUSES.map((status) => ({
    id: status,
    label: statusLabel(status),
    tone: status === 'Définitif' ? 'success' : status === 'ValidéSemaine' ? 'warn' : 'muted'
  }))
)

const columns = computed(() => {
  const cols = [{ key: 'month', label: t('cra.col_period') }]
  if (canValidateCra.value) {
    cols.push({ key: 'user', label: t('cra.col_user') })
  }
  cols.push(
    { key: 'client', label: t('cra.col_client') },
    { key: 'mission', label: t('cra.col_mission') },
    { key: 'hours', label: t('cra.col_hours') },
    { key: 'status', label: t('cra.col_status') },
    { key: 'updatedAt', label: t('cra.col_updated') },
    { key: 'actions', label: '' }
  )
  return cols
})

const kpi = computed(() => {
  const items = rawItems.value
  const key = currentMonthKey()
  const current = items.find((ts) => ts.month === key)
  const [y, m] = key.split('-').map(Number)
  const currentMonthLabel = new Date(y, m - 1, 1).toLocaleDateString(
    locale.value === 'en' ? 'en-US' : 'fr-FR',
    { month: 'long', year: 'numeric' }
  )
  return {
    total: items.length,
    drafts: countCraByStatus(items, 'Brouillon'),
    submitted: countCraByStatus(items, 'ValidéSemaine'),
    finalized: countCraByStatus(items, 'Définitif'),
    currentStatusLabel: current ? statusLabel(current.status) : '—',
    currentMonthLabel
  }
})

const formatMonth = (raw: string) => {
  const [y, m] = raw.split('-').map(Number)
  return new Date(y, m - 1, 1).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    month: 'long',
    year: 'numeric'
  })
}

const formatUpdated = (raw: string) => {
  if (!raw) return '—'
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString(locale.value === 'en' ? 'en-US' : 'fr-FR', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

const openMonth = async (monthKey: string, opts?: { fromModal?: boolean }) => {
  creating.value = true
  errorMsg.value = ''
  periodModalError.value = ''
  try {
    const res = await apiFetch<{
      data?: { id?: string; created?: boolean }
      id?: string
      created?: boolean
    }>(`/api/cra/timesheets?month=${encodeURIComponent(monthKey)}`)
    const ts = res?.data ?? res
    if (ts?.id) {
      const created = Boolean(ts.created)
      const notice = created ? 'created' : 'existing'
      periodModalOpen.value = false
      await navigateTo({ path: `/cra/${ts.id}`, query: { notice } })
      return
    }
    await refresh()
  } catch (err) {
    const msg = mapCraError(err, t('cra.open_error'))
    if (opts?.fromModal) {
      periodModalError.value = msg
    } else {
      errorMsg.value = msg
    }
  } finally {
    creating.value = false
  }
}

const openCurrentMonth = () => openMonth(currentMonthKey())

const openPeriodModal = () => {
  const clamped = clampCreatePeriodSelection(periodForm.year, periodForm.month)
  periodForm.year = clamped.year
  periodForm.month = clamped.month
  periodModalError.value = ''
  periodModalOpen.value = true
}

const confirmPeriodModal = async () => {
  const clamped = clampCreatePeriodSelection(periodForm.year, periodForm.month)
  periodForm.year = clamped.year
  periodForm.month = clamped.month
  if (!periodMonthOptions.value.length) {
    periodModalError.value = t('cra.new_period_invalid')
    return
  }
  await openMonth(monthKeyFromParts(periodForm.year, periodForm.month), { fromModal: true })
}

const onTimesheetAdminChange = async (action: 'unvalidate' | 'delete') => {
  errorMsg.value = ''
  switch (action) {
    case 'delete':
      successMsg.value = t('cra.delete_ok')
      break
    case 'unvalidate':
      successMsg.value = t('cra.unvalidate_ok')
      break
    default: {
      const _exhaustive: never = action
      return _exhaustive
    }
  }
  await refresh()
}

const onTimesheetDeleteError = (message: string) => {
  successMsg.value = ''
  errorMsg.value = message
}
</script>

<style scoped>
.cra-table-wrap { overflow: hidden; }

.cra-month {
  font-weight: 600;
  color: var(--kore-text);
}

.cra-user {
  font-weight: 500;
  color: var(--kore-text);
}

.cra-link {
  font-weight: 500;
  color: var(--kore-brand-blue);
  text-decoration: none;
}

.cra-link:hover {
  text-decoration: underline;
}

.cra-link--truncate {
  max-width: 14rem;
  display: inline-block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.cra-context {
  color: var(--kore-text);
  max-width: 14rem;
  display: inline-block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.cra-context--empty {
  color: var(--kore-text-muted);
  font-style: italic;
}

.cra-hours {
  font-variant-numeric: tabular-nums;
  color: var(--kore-text);
}

.cra-updated {
  color: var(--kore-text-muted);
  font-size: var(--kore-text-caption);
  white-space: nowrap;
}

.cra-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}

.muted { color: var(--kore-text-muted); }

.cra-empty-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.cra-period-form {
  display: grid;
  gap: var(--kore-space-md);
}

.cra-period-form__title {
  margin: 0;
  font-size: var(--kore-text-h3);
  color: var(--kore-text);
}

.cra-period-form__hint {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.cra-period-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--kore-space-sm);
}

.cra-period-form__field {
  display: grid;
  gap: var(--kore-space-xs);
}

.cra-period-form__field label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.cra-period-form__field select {
  width: 100%;
  min-height: 2.5rem;
  padding: var(--kore-space-xs) var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-surface);
  color: var(--kore-text);
}

.cra-period-form__error {
  margin: 0;
  color: var(--kore-error);
  font-size: var(--kore-text-small);
}

.cra-period-form__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}

.flash {
  margin-top: var(--kore-space-md);
  font-size: var(--kore-text-small);
}

.flash--error { color: var(--kore-error); }

.cra-kanban-card {
  display: grid;
  gap: var(--kore-space-xs);
}

.cra-kanban-card__title {
  margin: 0;
  font-weight: 600;
  color: var(--kore-text);
}

.cra-kanban-card__meta {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

@media (max-width: 768px) {
  .cra-link--truncate,
  .cra-context {
    max-width: 8rem;
  }

  .cra-actions :deep(.app-btn),
  .cra-kanban-card :deep(.app-btn),
  .cra-empty-actions :deep(.app-btn),
  .cra-period-form__actions :deep(.app-btn) {
    width: 100%;
  }

  .cra-period-form__row {
    grid-template-columns: 1fr;
  }

  .cra-period-form__actions {
    flex-direction: column-reverse;
  }
}
</style>
