<template>
  <div class="wf-page">
    <AppPageHeader :title="$t('workflows.title')" :subtitle="$t('workflows.subtitle')">
      <template #actions>
        <AppButton
          v-if="guideRef?.dismissed"
          variant="ghost"
          size="sm"
          type="button"
          @click="guideRef?.showAgain()"
        >
          {{ $t('guides.show') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppSectionGuide ref="guideRef" guide-key="admin.workflows" />

    <AppCard padding="lg" class="mb">
      <fieldset class="wf-preset">
        <legend class="wf-section-title">{{ $t('workflows.preset_label') }}</legend>
        <div class="wf-preset-options" role="radiogroup" :aria-label="$t('workflows.preset_label')">
          <label
            v-for="code in WORKFLOW_PRESET_CODES"
            :key="code"
            class="wf-preset-option"
            :class="{ 'wf-preset-option--active': selectedPreset === code }"
          >
            <input
              v-model="selectedPreset"
              class="wf-preset-option__input"
              type="radio"
              name="wf-preset"
              :value="code"
              :disabled="loading"
            />
            <span class="wf-preset-option__title">{{ $t(WORKFLOW_PRESETS[code].labelKey) }}</span>
            <span class="wf-preset-option__desc">{{ $t(WORKFLOW_PRESETS[code].descKey) }}</span>
          </label>
        </div>
      </fieldset>
    </AppCard>

    <p v-if="errorMsg" class="flash flash--error" role="alert">{{ errorMsg }}</p>
    <p v-if="flash" class="flash" role="status">{{ flash }}</p>

    <template v-if="editor">
      <AppCard padding="lg" class="mb">
        <div class="settings-howto" role="note">
          <p class="settings-howto__title">{{ $t('workflows.howto.title') }}</p>
          <ol class="settings-howto__list settings-howto__list--ordered">
            <li>{{ $t('workflows.howto.step_states') }}</li>
            <li>{{ $t('workflows.howto.step_transitions') }}</li>
            <li>{{ $t('workflows.howto.step_roles') }}</li>
            <li>{{ $t('workflows.howto.step_save') }}</li>
            <li>{{ $t(presetHowtoKey) }}</li>
          </ol>
        </div>

        <p class="settings-hint">
          <strong>{{ $t('workflows.entity_type') }}:</strong> {{ editor.entityType }}
        </p>
      </AppCard>

      <AppCard padding="lg" class="mb">
        <h2 class="wf-section-title">{{ $t('workflows.diagram_title') }}</h2>
        <WorkflowDiagram :states="editor.states" :transitions="editor.transitions" />
      </AppCard>

      <AppCard padding="lg" class="mb">
        <h2 class="wf-section-title">{{ $t('workflows.states_title') }}</h2>
        <WorkflowStateForm
          :states="editor.states"
          :transitions="editor.transitions"
          lock-codes
          @update:states="editor.states = $event"
        />
      </AppCard>

      <AppCard padding="lg" class="mb">
        <h2 class="wf-section-title">{{ $t('workflows.transitions_title') }}</h2>
        <WorkflowTransitionForm
          :states="editor.states"
          :transitions="editor.transitions"
          @update:transitions="editor.transitions = $event"
        />
      </AppCard>

      <AppCard padding="lg">
        <ul v-if="validationErrors.length" class="wf-validation" role="alert">
          <li v-for="(err, index) in validationErrors" :key="index">{{ err }}</li>
        </ul>
        <AppButton variant="primary" :loading="saving" @click="saveWorkflow">
          {{ $t('workflows.save') }}
        </AppButton>
      </AppCard>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { WorkflowDefinition, WorkflowPresetCode } from '~/composables/useWorkflowDefinition'
import {
  WORKFLOW_PRESET_CODES,
  WORKFLOW_PRESETS,
  buildPayload,
  buildPresetDefinition,
  normalizeDefinition,
  validateDefinition
} from '~/composables/useWorkflowDefinition'

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const { extractFetchError } = useApiError()

const guideRef = ref<{ showAgain: () => void; dismissed: boolean } | null>(null)
const selectedPreset = ref<WorkflowPresetCode>('leave.request')
const loading = ref(false)
const saving = ref(false)
const errorMsg = ref('')
const flash = ref('')
const editor = ref<WorkflowDefinition | null>(null)
const isHydrating = ref(true)

const presetHowtoKey = computed(() => WORKFLOW_PRESETS[selectedPreset.value].howtoKey)

const validationErrors = computed(() =>
  editor.value
    ? validateDefinition(editor.value).map((code) => t(`workflows.validation.${code}`))
    : []
)

const loadWorkflow = async (code: WorkflowPresetCode) => {
  loading.value = true
  errorMsg.value = ''
  flash.value = ''
  try {
    const res = await $fetch<{ data?: Parameters<typeof normalizeDefinition>[0] }>(
      `/api/admin/workflows/${encodeURIComponent(code)}`
    )
    const raw = (res?.data ?? res) as Parameters<typeof normalizeDefinition>[0]
    editor.value = normalizeDefinition(raw, code)
    if (!editor.value.entityType) {
      editor.value.entityType = WORKFLOW_PRESETS[code].entityType
    }
  } catch (e) {
    const statusCode = e && typeof e === 'object' && 'statusCode' in e ? (e as { statusCode?: number }).statusCode : undefined
    if (statusCode === 404) {
      editor.value = buildPresetDefinition(code)
      flash.value = t('workflows.not_found_preset')
      errorMsg.value = ''
    } else {
      editor.value = null
      errorMsg.value = extractFetchError(e, t('workflows.error_load'))
    }
  } finally {
    loading.value = false
    isHydrating.value = false
  }
}

watch(selectedPreset, (code) => {
  if (isHydrating.value) return
  loadWorkflow(code)
})

const saveWorkflow = async () => {
  if (!editor.value) return
  const errors = validateDefinition(editor.value)
  if (errors.length) {
    errorMsg.value = validationErrors.value[0] ?? t('common.error')
    return
  }

  saving.value = true
  errorMsg.value = ''
  flash.value = ''
  try {
    await $fetch('/api/admin/workflows', {
      method: 'POST',
      body: buildPayload(editor.value)
    })
    flash.value = t('workflows.saved')
    await loadWorkflow(selectedPreset.value)
  } catch (e) {
    errorMsg.value = extractFetchError(e)
  } finally {
    saving.value = false
  }
}

await loadWorkflow(selectedPreset.value)
</script>

<style scoped>
.mb { margin-bottom: var(--kore-space-lg); }

.wf-section-title {
  margin: 0 0 var(--kore-space-md);
  font-size: var(--kore-text-h3);
}

.wf-preset {
  border: none;
  padding: 0;
  margin: 0;
}

.wf-preset-options {
  display: grid;
  gap: var(--kore-space-sm);
  max-width: var(--kore-form-wide-max);
}

.wf-preset-option {
  display: grid;
  gap: 0.2rem;
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  cursor: pointer;
}

.wf-preset-option--active {
  border-color: var(--kore-primary);
  background: var(--kore-bg-elevated);
}

.wf-preset-option__input {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.wf-preset-option__title {
  font-size: var(--kore-text-small);
  font-weight: 600;
}

.wf-preset-option__desc {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.settings-howto {
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
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

.settings-howto__list--ordered {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.settings-hint {
  margin: 0;
  color: var(--kore-text-muted);
  font-size: var(--kore-text-small);
}

.wf-validation {
  margin: 0 0 var(--kore-space-md);
  padding-left: 1.25rem;
  color: var(--kore-status-danger);
  font-size: var(--kore-text-small);
}

.flash {
  margin-bottom: var(--kore-space-md);
  font-size: var(--kore-text-small);
}

.flash--error {
  color: var(--kore-status-danger);
}

@media (max-width: 768px) {
  .wf-page :deep(.app-button) {
    width: 100%;
  }
}
</style>
