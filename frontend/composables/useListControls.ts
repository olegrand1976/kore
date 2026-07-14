import { computed, onMounted, reactive, ref, toValue, watch, type MaybeRefOrGetter, type Ref } from 'vue'

export type SortDir = 'asc' | 'desc'
export type ListView = 'table' | 'kanban'
export type SortValueType = 'string' | 'number' | 'date'

export type SortKeyDef<T> = {
  key: string
  label: string
  type: SortValueType
  accessor: (item: T) => string | number | null | undefined
}

export type FilterSelectOption = {
  value: string
  label: string
}

export type FilterDef<T> =
  | {
      type: 'select'
      label: string
      options: FilterSelectOption[]
      defaultValue?: string
      match: (item: T, value: string) => boolean
    }
  | {
      type: 'search'
      label: string
      placeholder?: string
      defaultValue?: string
      match: (item: T, query: string) => boolean
    }
  | {
      type: 'month'
      label: string
      defaultValue?: string
      match: (item: T, value: string) => boolean
    }

export type ListControlsOptions<T> = {
  storageKey?: string
  defaultSort?: { key: string; dir: SortDir }
  defaultView?: ListView
  kanbanEnabled?: boolean
  filters?: MaybeRefOrGetter<Record<string, FilterDef<T>> | undefined>
  sortKeys: MaybeRefOrGetter<SortKeyDef<T>[]>
}

type StoredState = {
  filters: Record<string, string>
  sortKey: string
  sortDir: SortDir
  view: ListView
}

function normalizeString(value: unknown): string {
  if (value == null) return ''
  return String(value).trim().toLowerCase()
}

export function applyTextSearch(query: string, ...fields: unknown[]): boolean {
  const needle = normalizeString(query)
  if (!needle) return true
  return fields.some((field) => normalizeString(field).includes(needle))
}

export function compareValues(
  a: string | number | null | undefined,
  b: string | number | null | undefined,
  type: SortValueType
): number {
  if (a == null && b == null) return 0
  if (a == null) return 1
  if (b == null) return -1

  if (type === 'number') {
    const na = Number(a)
    const nb = Number(b)
    if (Number.isNaN(na) && Number.isNaN(nb)) return 0
    if (Number.isNaN(na)) return 1
    if (Number.isNaN(nb)) return -1
    return na - nb
  }

  if (type === 'date') {
    const da = new Date(String(a)).getTime()
    const db = new Date(String(b)).getTime()
    if (Number.isNaN(da) && Number.isNaN(db)) return 0
    if (Number.isNaN(da)) return 1
    if (Number.isNaN(db)) return -1
    return da - db
  }

  return normalizeString(a).localeCompare(normalizeString(b), undefined, { sensitivity: 'base' })
}

export function groupByKey<T>(items: T[], keyFn: (item: T) => string): Record<string, T[]> {
  const groups: Record<string, T[]> = {}
  for (const item of items) {
    const key = keyFn(item) || 'unknown'
    if (!groups[key]) groups[key] = []
    groups[key].push(item)
  }
  return groups
}

function defaultFilterValues<T>(filters?: Record<string, FilterDef<T>>): Record<string, string> {
  const values: Record<string, string> = {}
  if (!filters) return values
  for (const [key, def] of Object.entries(filters)) {
    values[key] = def.defaultValue ?? (def.type === 'select' ? '' : '')
  }
  return values
}

function readStoredState(storageKey: string | undefined): Partial<StoredState> | null {
  if (!import.meta.client || !storageKey) return null
  try {
    const raw = localStorage.getItem(`kore-list:${storageKey}`)
    if (!raw) return null
    return JSON.parse(raw) as Partial<StoredState>
  } catch {
    return null
  }
}

function writeStoredState(storageKey: string | undefined, state: StoredState) {
  if (!import.meta.client || !storageKey) return
  try {
    localStorage.setItem(`kore-list:${storageKey}`, JSON.stringify(state))
  } catch {
    // ignore quota errors
  }
}

export function useListControls<T>(items: Ref<T[]>, options: ListControlsOptions<T>) {
  const defaultSortKey = options.defaultSort?.key ?? toValue(options.sortKeys)?.[0]?.key ?? ''
  const defaultSortDir = options.defaultSort?.dir ?? 'asc'

  const filterDefs = computed(() => toValue(options.filters))

  const stored = readStoredState(options.storageKey)
  const initialFilters = {
    ...defaultFilterValues(filterDefs.value),
    ...(stored?.filters ?? {})
  }

  const filterValues = reactive<Record<string, string>>(initialFilters)
  const sortKey = ref(
    stored?.sortKey && toValue(options.sortKeys).some((s) => s.key === stored.sortKey)
      ? stored.sortKey
      : defaultSortKey
  )
  const sortDir = ref<SortDir>(stored?.sortDir === 'desc' || stored?.sortDir === 'asc' ? stored.sortDir : defaultSortDir)
  const view = ref<ListView>(
    options.kanbanEnabled && (stored?.view === 'kanban' || stored?.view === 'table')
      ? stored.view
      : (options.defaultView ?? 'table')
  )

  const activeSort = computed(
    () => toValue(options.sortKeys).find((s) => s.key === sortKey.value) ?? toValue(options.sortKeys)[0]
  )

  const filteredItems = computed(() => {
    let result = items.value
    const filters = filterDefs.value
    if (!filters) return result

    for (const [key, def] of Object.entries(filters)) {
      const value = filterValues[key] ?? ''
      if (!value) continue
      if (def.type === 'select') {
        result = result.filter((item) => def.match(item, value))
      } else if (def.type === 'search') {
        result = result.filter((item) => def.match(item, value))
      } else if (def.type === 'month') {
        result = result.filter((item) => def.match(item, value))
      } else {
        const _exhaustive: never = def
        void _exhaustive
      }
    }
    return result
  })

  const sortedItems = computed(() => {
    const sort = activeSort.value
    if (!sort) return [...filteredItems.value]
    const dir = sortDir.value === 'asc' ? 1 : -1
    return [...filteredItems.value].sort((a, b) => {
      const cmp = compareValues(sort.accessor(a), sort.accessor(b), sort.type)
      return cmp * dir
    })
  })

  const hasActiveFilters = computed(() => {
    const defaults = defaultFilterValues(filterDefs.value)
    return Object.entries(filterValues).some(([key, value]) => (value ?? '') !== (defaults[key] ?? ''))
  })

  function setFilter(key: string, value: string) {
    filterValues[key] = value
  }

  function setSort(key: string, dir?: SortDir) {
    sortKey.value = key
    if (dir) sortDir.value = dir
  }

  function setSortDir(dir: SortDir) {
    sortDir.value = dir
  }

  function setView(next: ListView) {
    if (next === 'kanban' && !options.kanbanEnabled) return
    view.value = next
  }

  function resetFilters() {
    const defaults = defaultFilterValues(filterDefs.value)
    for (const key of Object.keys(defaults)) {
      filterValues[key] = defaults[key] ?? ''
    }
  }

  function resetAll() {
    resetFilters()
    sortKey.value = defaultSortKey
    sortDir.value = defaultSortDir
    view.value = options.defaultView ?? 'table'
  }

  watch(
    [filterValues, sortKey, sortDir, view],
    () => {
      writeStoredState(options.storageKey, {
        filters: { ...filterValues },
        sortKey: sortKey.value,
        sortDir: sortDir.value,
        view: view.value
      })
    },
    { deep: true }
  )

  return {
    filterValues,
    filterDefs,
    sortKey,
    sortDir,
    view,
    filteredItems,
    sortedItems,
    hasActiveFilters,
    activeSort,
    setFilter,
    setSort,
    setSortDir,
    setView,
    resetFilters,
    resetAll
  }
}

/** Aligne une ref `month` (YYYY-MM) avec le filtre mois persisté, puis recharge les données. */
export function syncListMonthFilter(
  filterValues: Record<string, string>,
  month: Ref<string>,
  refresh: () => Promise<void>
) {
  onMounted(async () => {
    const stored = filterValues.month?.trim()
    if (stored && stored !== month.value) {
      month.value = stored
      await refresh()
      return
    }
    if (!filterValues.month) {
      filterValues.month = month.value
    }
  })
}
