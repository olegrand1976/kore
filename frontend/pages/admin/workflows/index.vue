<template>
  <div>
    <AppPageHeader :title="$t('workflows.title')" />

    <AppCard padding="lg" class="mb">
      <form class="load-form" @submit.prevent="loadWorkflow">
        <AppInput v-model="workflowCode" :label="$t('workflows.code_label')" placeholder="leave.request" required />
        <AppButton variant="primary" size="sm" type="submit" :disabled="loading">
          {{ $t('workflows.load') }}
        </AppButton>
      </form>
    </AppCard>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <p v-if="flash" class="flash" role="status">{{ flash }}</p>

    <template v-if="definition">
      <AppCard padding="lg" class="mb">
        <h2 class="section-title">{{ $t('workflows.states') }}</h2>
        <ul class="wf-list">
          <li v-for="state in definition.states" :key="String(state.code ?? state.Code)">
            <AppBadge variant="neutral">{{ state.label ?? state.Label ?? state.code ?? state.Code }}</AppBadge>
            <span class="wf-meta">
              {{ state.code ?? state.Code }}
              <template v-if="state.isInitial ?? state.IsInitial"> · {{ $t('workflows.initial') }}</template>
              <template v-if="state.isFinal ?? state.IsFinal"> · {{ $t('workflows.final') }}</template>
            </span>
          </li>
        </ul>
      </AppCard>

      <AppCard padding="lg" class="mb">
        <h2 class="section-title">{{ $t('workflows.transitions') }}</h2>
        <AppTable :columns="transitionColumns" :rows="transitionRows" row-key="key" />
      </AppCard>

      <AppCard padding="lg">
        <h2 class="section-title">{{ $t('workflows.json_editor') }}</h2>
        <textarea v-model="jsonEditor" class="json-editor" rows="16" :aria-label="$t('workflows.json_editor')" />
        <div class="json-actions">
          <AppButton variant="primary" size="sm" :disabled="saving" @click="saveWorkflow">
            {{ $t('workflows.save') }}
          </AppButton>
        </div>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
type WorkflowState = {
  code?: string
  Code?: string
  label?: string
  Label?: string
  isInitial?: boolean
  IsInitial?: boolean
  isFinal?: boolean
  IsFinal?: boolean
}

type WorkflowTransition = {
  from?: string
  From?: string
  to?: string
  To?: string
  action?: string
  Action?: string
  guard?: string
  Guard?: string
}

type WorkflowDefinition = {
  code?: string
  Code?: string
  entityType?: string
  EntityType?: string
  states?: WorkflowState[]
  States?: WorkflowState[]
  transitions?: WorkflowTransition[]
  Transitions?: WorkflowTransition[]
}

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const { extractFetchError } = useApiError()

const workflowCode = ref('leave.request')
const loading = ref(false)
const saving = ref(false)
const errorMsg = ref('')
const flash = ref('')
const definition = ref<WorkflowDefinition | null>(null)
const jsonEditor = ref('')

const transitionColumns = computed(() => [
  { key: 'from', label: t('workflows.col_from') },
  { key: 'to', label: t('workflows.col_to') },
  { key: 'action', label: t('workflows.col_action') }
])

const transitionRows = computed(() => {
  const transitions = definition.value?.transitions ?? definition.value?.Transitions ?? []
  return transitions.map((tr, index) => ({
    key: String(index),
    from: tr.from ?? tr.From ?? '',
    to: tr.to ?? tr.To ?? '',
    action: tr.action ?? tr.Action ?? ''
  }))
})

const normalizeDefinition = (raw: WorkflowDefinition) => ({
  code: raw.code ?? raw.Code ?? workflowCode.value,
  entityType: raw.entityType ?? raw.EntityType ?? '',
  states: (raw.states ?? raw.States ?? []).map((s) => ({
    code: s.code ?? s.Code ?? '',
    label: s.label ?? s.Label ?? '',
    isInitial: s.isInitial ?? s.IsInitial ?? false,
    isFinal: s.isFinal ?? s.IsFinal ?? false
  })),
  transitions: (raw.transitions ?? raw.Transitions ?? []).map((tr) => ({
    from: tr.from ?? tr.From ?? '',
    to: tr.to ?? tr.To ?? '',
    action: tr.action ?? tr.Action ?? '',
    guard: tr.guard ?? tr.Guard ?? ''
  }))
})

const loadWorkflow = async () => {
  loading.value = true
  errorMsg.value = ''
  flash.value = ''
  try {
    const res = await $fetch<{ data?: WorkflowDefinition }>(`/api/admin/workflows/${encodeURIComponent(workflowCode.value)}`)
    const raw = (res?.data ?? res) as WorkflowDefinition
    definition.value = raw
    jsonEditor.value = JSON.stringify(normalizeDefinition(raw), null, 2)
  } catch (e) {
    definition.value = null
    errorMsg.value = extractFetchError(e)
  } finally {
    loading.value = false
  }
}

const saveWorkflow = async () => {
  saving.value = true
  errorMsg.value = ''
  flash.value = ''
  try {
    const payload = JSON.parse(jsonEditor.value) as WorkflowDefinition
    await $fetch('/api/admin/workflows', { method: 'POST', body: payload })
    flash.value = t('workflows.saved')
    await loadWorkflow()
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    saving.value = false
  }
}

await loadWorkflow()
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }
.load-form {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
  align-items: flex-end;
  max-width: var(--kore-form-wide-max);
}
.section-title { margin: 0 0 var(--kore-space-md); font-size: var(--kore-text-h3); }
.wf-list { list-style: none; margin: 0; padding: 0; display: grid; gap: var(--kore-space-sm); }
.wf-meta { margin-left: var(--kore-space-sm); font-size: var(--kore-text-small); color: var(--kore-text-muted); }
.json-editor {
  width: 100%;
  font-family: ui-monospace, monospace;
  font-size: var(--kore-text-small);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-sm);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
  resize: vertical;
}
.json-actions { margin-top: var(--kore-space-md); }
.flash { margin-bottom: var(--kore-space-md); font-size: var(--kore-text-small); }
.flash--error { color: var(--kore-status-danger); }
@media (max-width: 768px) {
  .load-form :deep(.app-button),
  .json-actions :deep(.app-button) { width: 100%; }
}
</style>
