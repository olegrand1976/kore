<template>
  <div>
    <AppPageHeader :title="$t('notifications.title')">
      <template #actions>
        <AppButton variant="primary" size="sm" @click="openCreate">
          {{ $t('notifications.add_rule') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard padding="lg" class="mb">
      <h3 class="section-title">{{ $t('notifications.rules') }}</h3>
      <AppListToolbar
        :filters="ruleListFilters"
        :filter-values="ruleFilterValues"
        :sort-keys="ruleSortKeys"
        :sort-key="ruleSortKey"
        :sort-dir="ruleSortDir"
        :has-active-filters="ruleHasActiveFilters"
        @update:filter="setRuleFilter"
        @update:sort-key="setRuleSort($event)"
        @update:sort-dir="setRuleSortDir"
        @reset="resetRuleFilters"
      />
      <AppTable
        :columns="ruleColumns"
        :rows="ruleDisplayRows"
        :loading="rulesPending"
        :empty-title="ruleHasActiveFilters ? $t('common.list.no_results') : $t('notifications.rules_empty')"
        row-key="code"
      >
        <template #cell-frequency="{ value }">
          {{ frequencyLabel(value) }}
        </template>
        <template #cell-actions="{ row }">
          <AppButton variant="ghost" size="sm" @click="openEdit(row)">
            {{ $t('common.edit') }}
          </AppButton>
        </template>
      </AppTable>
    </AppCard>

    <AppCard v-if="showForm" padding="lg" class="mb settings-form">
      <h3 class="section-title">
        {{ editingCode ? $t('notifications.edit_title') : $t('notifications.add_title') }}
      </h3>
      <div class="settings-howto" role="note">
        <p class="settings-howto__title">{{ $t('notifications.howto.title') }}</p>
        <ul class="settings-howto__list">
          <li>{{ $t('notifications.howto.item_code') }}</li>
          <li>{{ $t('notifications.howto.item_trigger') }}</li>
          <li>{{ $t('notifications.howto.item_frequency') }}</li>
          <li>{{ $t('notifications.howto.item_template') }}</li>
        </ul>
      </div>
      <form class="settings-form__grid" @submit.prevent="save">
        <AppInput
          id="rule-code"
          v-model="form.code"
          :disabled="!!editingCode"
          required
        >
          <template #label>
            <span class="settings-labelRow">
              <span>{{ $t('notifications.col_code') }}</span>
              <AppTooltip :button-label="$t('common.info')">
                {{ $t('notifications.tooltip.code') }}
              </AppTooltip>
            </span>
          </template>
        </AppInput>
        <p class="settings-hint settings-hint--tight">
          {{ $t('notifications.hint.code') }}
        </p>
        <AppInput
          id="rule-trigger"
          v-model="form.trigger"
          required
        >
          <template #label>
            <span class="settings-labelRow">
              <span>{{ $t('notifications.col_trigger') }}</span>
              <AppTooltip :button-label="$t('common.info')">
                {{ $t('notifications.tooltip.trigger') }}
              </AppTooltip>
            </span>
          </template>
        </AppInput>
        <p class="settings-hint settings-hint--tight">
          {{ $t('notifications.hint.trigger') }}
        </p>
        <div class="settings-field">
          <label for="rule-frequency" class="settings-labelRow">
            <span>{{ $t('notifications.col_frequency') }}</span>
            <AppTooltip :button-label="$t('common.info')">
              {{ $t('notifications.tooltip.frequency') }}
            </AppTooltip>
          </label>
          <select id="rule-frequency" v-model="form.frequency" required>
            <option v-for="f in frequencies" :key="f" :value="f">
              {{ frequencyLabel(f) }}
            </option>
          </select>
          <p class="settings-hint">{{ $t('notifications.hint.frequency') }}</p>
        </div>
        <div class="settings-field settings-field--full">
          <label for="rule-template" class="settings-labelRow">
            <span>{{ $t('notifications.col_template') }}</span>
            <AppTooltip :button-label="$t('common.info')">
              {{ $t('notifications.tooltip.template') }}
            </AppTooltip>
          </label>
          <textarea
            id="rule-template"
            v-model="form.template"
            rows="4"
            required
          />
          <p class="settings-hint">{{ $t('notifications.hint.template') }}</p>
        </div>
        <label class="settings-toggle">
          <input v-model="form.attachPdf" type="checkbox" />
          <span class="settings-labelRow">
            <span>{{ $t('notifications.attach_pdf') }}</span>
            <AppTooltip :button-label="$t('common.info')">
              {{ $t('notifications.tooltip.attach_pdf') }}
            </AppTooltip>
          </span>
        </label>
        <p class="settings-hint settings-hint--tight">
          {{ $t('notifications.hint.attach_pdf') }}
        </p>
        <div class="settings-form__actions">
          <AppButton variant="ghost" size="sm" type="button" @click="closeForm">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" size="sm" type="submit" :disabled="saving">
            {{ $t('common.save') }}
          </AppButton>
        </div>
      </form>
      <p v-if="formError" class="settings-flash settings-flash--error" role="alert">{{ formError }}</p>
    </AppCard>

    <AppCard padding="lg">
      <h3 class="section-title">{{ $t('notifications.journal') }}</h3>
      <AppTable
        :columns="journalColumns"
        :rows="journalRows"
        :loading="journalPending"
        :empty-title="$t('notifications.journal_empty')"
      />
    </AppCard>

    <p v-if="flash" class="settings-flash" :class="{ 'settings-flash--error': flashError }" role="status">
      {{ flash }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { useListControls } from '~/composables/useListControls'

type NotificationRule = {
  id?: string
  ID?: string
  code?: string
  Code?: string
  trigger?: string
  Trigger?: string
  frequency?: string
  Frequency?: string
  template?: string
  Template?: string
  attachPdf?: boolean
  AttachPDF?: boolean
}

type RuleRow = {
  id: string
  code: string
  trigger: string
  frequency: string
  template: string
  attachPdf: boolean
}

const FREQUENCIES = [
  'immediate',
  'morning',
  'monday',
  'friday',
  'last_monday_of_month'
] as const

type Frequency = (typeof FREQUENCIES)[number]

definePageMeta({ layout: 'default', middleware: ['admin'] })

const { t } = useI18n()
const { extractFetchError } = useApiError()

const { data: rulesData, pending: rulesPending, refresh: refreshRules } = await useFetch('/api/notifications/rules')
const { data: journalData, pending: journalPending } = await useFetch('/api/notifications/journal')

const showForm = ref(false)
const saving = ref(false)
const editingCode = ref('')
const formError = ref('')
const flash = ref('')
const flashError = ref(false)

const form = reactive({
  code: '',
  trigger: '',
  frequency: 'immediate' as Frequency,
  template: '',
  attachPdf: false
})

const frequencies = FREQUENCIES

const frequencyLabel = (value: string) => {
  const key = `notifications.frequency.${value}`
  const translated = t(key)
  return translated === key ? value : translated
}

const ruleColumns = computed(() => [
  { key: 'code', label: t('notifications.col_code') },
  { key: 'trigger', label: t('notifications.col_trigger') },
  { key: 'frequency', label: t('notifications.col_frequency') },
  { key: 'actions', label: '' }
])

const journalColumns = computed(() => [
  { key: 'subject', label: t('notifications.col_subject') },
  { key: 'status', label: t('notifications.col_status') }
])

const toRuleRow = (r: NotificationRule): RuleRow => ({
  id: r.id ?? r.ID ?? '',
  code: r.code ?? r.Code ?? '',
  trigger: r.trigger ?? r.Trigger ?? '',
  frequency: r.frequency ?? r.Frequency ?? '',
  template: r.template ?? r.Template ?? '',
  attachPdf: r.attachPdf ?? r.AttachPDF ?? false
})

const ruleRows = computed(() => {
  const items = (rulesData.value as { data?: NotificationRule[] })?.data ?? []
  return items.map(toRuleRow)
})

const ruleListFilters = computed(() => ({
  frequency: {
    type: 'select' as const,
    label: t('notifications.col_frequency'),
    options: FREQUENCIES.map((freq) => ({
      value: freq,
      label: frequencyLabel(freq)
    })),
    match: (row: RuleRow, value: string) => row.frequency === value
  }
}))

const ruleSortKeys = computed(() => [
  { key: 'code', label: t('notifications.col_code'), type: 'string' as const, accessor: (row: RuleRow) => row.code },
  { key: 'trigger', label: t('notifications.col_trigger'), type: 'string' as const, accessor: (row: RuleRow) => row.trigger }
])

const {
  filterValues: ruleFilterValues,
  sortKey: ruleSortKey,
  sortDir: ruleSortDir,
  sortedItems: ruleSortedItems,
  hasActiveFilters: ruleHasActiveFilters,
  setFilter: setRuleFilter,
  setSort: setRuleSort,
  setSortDir: setRuleSortDir,
  resetFilters: resetRuleFilters
} = useListControls(ruleRows, {
  storageKey: 'notification-rules',
  defaultSort: { key: 'code', dir: 'asc' },
  filters: ruleListFilters,
  sortKeys: ruleSortKeys
})

const ruleDisplayRows = computed(() => ruleSortedItems.value)

const journalRows = computed(() => {
  const items = (journalData.value as { data?: Array<{ subject?: string; Subject?: string; status?: string; Status?: string }> })?.data ?? []
  return items.map((m) => ({
    subject: m.subject ?? m.Subject,
    status: m.status ?? m.Status
  }))
})

const resetForm = () => {
  form.code = ''
  form.trigger = ''
  form.frequency = 'immediate'
  form.template = ''
  form.attachPdf = false
  formError.value = ''
}

const openCreate = () => {
  editingCode.value = ''
  resetForm()
  showForm.value = true
}

const openEdit = (row: RuleRow) => {
  editingCode.value = row.code
  form.code = row.code
  form.trigger = row.trigger
  form.frequency = row.frequency as Frequency
  form.template = row.template
  form.attachPdf = row.attachPdf
  formError.value = ''
  showForm.value = true
}

const closeForm = () => {
  showForm.value = false
  editingCode.value = ''
  resetForm()
}

const save = async () => {
  saving.value = true
  formError.value = ''
  try {
    const payload = {
      code: form.code.trim(),
      trigger: form.trigger.trim(),
      frequency: form.frequency,
      template: form.template.trim(),
      attachPdf: form.attachPdf,
      recipientPolicy: {}
    }
    if (editingCode.value) {
      const row = ruleRows.value.find((r) => r.code === editingCode.value)
      await $fetch(`/api/notifications/rules/${row?.id || editingCode.value}`, {
        method: 'PUT',
        body: payload
      })
      flash.value = t('notifications.saved')
    } else {
      await $fetch('/api/notifications/rules', { method: 'POST', body: payload })
      flash.value = t('notifications.created')
    }
    flashError.value = false
    closeForm()
    await refreshRules()
  } catch (e) {
    formError.value = extractFetchError(e)
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }
.section-title { margin: 0 0 var(--kore-space-md); font-size: var(--kore-text-h3); }

.settings-form__grid {
  display: grid;
  gap: var(--kore-space-md);
  max-width: var(--kore-form-wide-max);
}

.settings-labelRow {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.settings-labelRow :deep(.app-tooltip__button) {
  width: 1.75rem;
  height: 1.75rem;
}

.settings-howto {
  margin: 0 0 var(--kore-space-lg);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  max-width: var(--kore-form-wide-max);
}

.settings-howto__title {
  margin: 0 0 var(--kore-space-sm);
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.settings-howto__list {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.settings-hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
  line-height: 1.35;
}

.settings-hint--tight {
  margin-top: calc(var(--kore-space-md) * -1);
}

.settings-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.settings-field--full {
  grid-column: 1 / -1;
}

.settings-field label,
.settings-field select,
.settings-field textarea {
  font-size: var(--kore-text-small);
}

.settings-field select,
.settings-field textarea {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg);
  color: var(--kore-text);
}

.settings-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--kore-text-small);
}

.settings-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
}

.settings-flash {
  margin-top: var(--kore-space-md);
  padding: 0.75rem 1rem;
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg-elevated);
  font-size: var(--kore-text-small);
}

.settings-flash--error {
  color: var(--kore-status-danger);
  border: 1px solid var(--kore-status-danger);
}

@media (max-width: 768px) {
  .settings-form__actions {
    flex-direction: column;
  }

  .settings-form__actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
