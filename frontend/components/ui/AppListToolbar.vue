<script setup lang="ts">
import type { FilterDef, ListView, SortDir, SortKeyDef } from '~/composables/useListControls'

const props = defineProps<{
  filters?: Record<string, FilterDef<unknown>>
  filterValues?: Record<string, string>
  sortKeys: SortKeyDef<unknown>[]
  sortKey: string
  sortDir: SortDir
  view?: ListView
  kanbanEnabled?: boolean
  hasActiveFilters?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:filter': [key: string, value: string]
  'update:sortKey': [value: string]
  'update:sortDir': [value: SortDir]
  'update:view': [value: ListView]
  reset: []
}>()

const { t } = useI18n()

const filterEntries = computed(() => Object.entries(props.filters ?? {}))

function onFilterChange(key: string, event: Event) {
  emit('update:filter', key, (event.target as HTMLInputElement | HTMLSelectElement).value)
}
</script>

<template>
  <AppCard padding="lg" class="list-toolbar" :class="{ 'list-toolbar--disabled': disabled }">
    <div class="list-toolbar__grid">
      <template v-for="[key, def] in filterEntries" :key="key">
        <AppInput
          v-if="def.type === 'search'"
          :id="`list-filter-${key}`"
          :model-value="filterValues?.[key] ?? ''"
          :label="def.label"
          :placeholder="def.placeholder"
          :disabled="disabled"
          @update:model-value="emit('update:filter', key, $event)"
        />
        <div v-else-if="def.type === 'month'" class="list-toolbar__field">
          <label :for="`list-filter-${key}`">{{ def.label }}</label>
          <input
            :id="`list-filter-${key}`"
            type="month"
            class="list-toolbar__month"
            :value="filterValues?.[key] ?? ''"
            :disabled="disabled"
            @change="onFilterChange(key, $event)"
          >
        </div>
        <div v-else-if="def.type === 'select'" class="list-toolbar__field">
          <label :for="`list-filter-${key}`">{{ def.label }}</label>
          <select
            :id="`list-filter-${key}`"
            class="list-toolbar__select"
            :value="filterValues?.[key] ?? ''"
            :disabled="disabled"
            @change="onFilterChange(key, $event)"
          >
            <option value="">{{ t('common.list.filter_all') }}</option>
            <option v-for="opt in def.options" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>
      </template>

      <div class="list-toolbar__field">
        <label for="list-sort-key">{{ t('common.list.sort') }}</label>
        <div class="list-toolbar__sort-row">
          <select
            id="list-sort-key"
            class="list-toolbar__select"
            :value="sortKey"
            :disabled="disabled"
            @change="emit('update:sortKey', ($event.target as HTMLSelectElement).value)"
          >
            <option v-for="opt in sortKeys" :key="opt.key" :value="opt.key">
              {{ opt.label }}
            </option>
          </select>
          <select
            id="list-sort-dir"
            class="list-toolbar__select list-toolbar__select--dir"
            :value="sortDir"
            :disabled="disabled"
            @change="emit('update:sortDir', ($event.target as HTMLSelectElement).value as SortDir)"
          >
            <option value="asc">{{ t('common.list.sort_asc') }}</option>
            <option value="desc">{{ t('common.list.sort_desc') }}</option>
          </select>
        </div>
      </div>

      <div v-if="kanbanEnabled" class="list-toolbar__field">
        <span class="list-toolbar__label">{{ t('common.list.view') }}</span>
        <div class="list-toolbar__view-toggle" role="group" :aria-label="t('common.list.view')">
          <button
            type="button"
            class="list-toolbar__view-btn"
            :class="{ 'list-toolbar__view-btn--active': view === 'table' }"
            :disabled="disabled"
            @click="emit('update:view', 'table')"
          >
            {{ t('common.list.view_table') }}
          </button>
          <button
            type="button"
            class="list-toolbar__view-btn"
            :class="{ 'list-toolbar__view-btn--active': view === 'kanban' }"
            :disabled="disabled"
            @click="emit('update:view', 'kanban')"
          >
            {{ t('common.list.view_kanban') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="hasActiveFilters && !disabled" class="list-toolbar__actions">
      <AppButton variant="ghost" size="sm" type="button" @click="emit('reset')">
        {{ t('common.list.reset_filters') }}
      </AppButton>
    </div>
  </AppCard>
</template>

<style scoped>
.list-toolbar {
  margin-bottom: var(--kore-space-lg);
}

.list-toolbar--disabled {
  opacity: 0.65;
  pointer-events: none;
}

.list-toolbar__select:disabled,
.list-toolbar__month:disabled {
  cursor: not-allowed;
}

.list-toolbar__view-btn:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.list-toolbar__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
  gap: var(--kore-space-md);
  align-items: end;
}

.list-toolbar__field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.list-toolbar__field label,
.list-toolbar__label {
  font-size: var(--kore-text-caption);
  font-weight: 600;
  color: var(--kore-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.list-toolbar__select,
.list-toolbar__month {
  width: 100%;
  padding: 0.55rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}

.list-toolbar__sort-row {
  display: flex;
  gap: var(--kore-space-sm);
}

.list-toolbar__select--dir {
  flex: 0 0 7rem;
}

.list-toolbar__view-toggle {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.list-toolbar__view-btn {
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  cursor: pointer;
}

.list-toolbar__view-btn--active {
  border-color: var(--kore-gold);
  background: var(--kore-bg-subtle);
  color: var(--kore-text);
  font-weight: 600;
}

.list-toolbar__actions {
  margin-top: var(--kore-space-md);
  padding-top: var(--kore-space-md);
  border-top: 1px solid var(--kore-border);
}

@media (max-width: 768px) {
  .list-toolbar__grid {
    grid-template-columns: 1fr;
  }

  .list-toolbar__sort-row {
    flex-direction: column;
  }

  .list-toolbar__select--dir {
    flex: 1 1 auto;
  }

  .list-toolbar__view-btn {
    flex: 1 1 calc(50% - 0.2rem);
    text-align: center;
  }
}
</style>
